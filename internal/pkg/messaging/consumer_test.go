package messaging

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
)

const releaseBody = `{"message_id":"b7a1c6f0-2f64-4f6e-9a49-0d5a0b1e6c11","type":"ReleaseStock","service_order_id":"01JD8Z4R3Q8M2V7K5N9XWABCDE","payload":{}}`

type fakeSQS struct {
	mu        sync.Mutex
	batches   [][]types.Message
	receiveEr error
	deleted   []string
}

func (f *fakeSQS) ReceiveMessage(ctx context.Context, _ *sqs.ReceiveMessageInput, _ ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	f.mu.Lock()
	if f.receiveEr != nil {
		err := f.receiveEr
		f.receiveEr = nil
		f.mu.Unlock()
		return nil, err
	}
	if len(f.batches) == 0 {
		f.mu.Unlock()
		<-ctx.Done()
		return nil, ctx.Err()
	}
	b := f.batches[0]
	f.batches = f.batches[1:]
	f.mu.Unlock()
	return &sqs.ReceiveMessageOutput{Messages: b}, nil
}

func (f *fakeSQS) DeleteMessage(_ context.Context, in *sqs.DeleteMessageInput, _ ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, aws.ToString(in.ReceiptHandle))
	return &sqs.DeleteMessageOutput{}, nil
}

func (f *fakeSQS) deletedHandles() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.deleted...)
}

func message(handle, body string) types.Message {
	return types.Message{ReceiptHandle: aws.String(handle), Body: aws.String(body)}
}

func newConsumer(client SQSClient, handlers map[string]Handler) *Consumer {
	return NewConsumer("https://sqs.test/queue", client, uow.NewInMemoryUoW(), NewMemoryDedupe(), handlers)
}

func TestHandle_RunsHandlerAndDeletes(t *testing.T) {
	var calls []Envelope
	client := &fakeSQS{}
	c := newConsumer(client, map[string]Handler{"ReleaseStock": func(_ context.Context, env Envelope) error {
		calls = append(calls, env)
		return nil
	}})

	c.handle(context.Background(), message("h1", releaseBody))

	require.Len(t, calls, 1)
	assert.Equal(t, "01JD8Z4R3Q8M2V7K5N9XWABCDE", calls[0].ServiceOrderID)
	assert.Equal(t, []string{"h1"}, client.deletedHandles())
}

func TestHandle_DuplicateIsDeletedWithoutRunningHandlerAgain(t *testing.T) {
	calls := 0
	client := &fakeSQS{}
	c := newConsumer(client, map[string]Handler{"ReleaseStock": func(context.Context, Envelope) error {
		calls++
		return nil
	}})

	c.handle(context.Background(), message("h1", releaseBody))
	c.handle(context.Background(), message("h2", releaseBody))

	assert.Equal(t, 1, calls)
	assert.Equal(t, []string{"h1", "h2"}, client.deletedHandles())
}

func TestHandle_UnknownTypeIsDeleted(t *testing.T) {
	client := &fakeSQS{}
	c := newConsumer(client, map[string]Handler{})

	c.handle(context.Background(), message("h1", `{"message_id":"b7a1c6f0-2f64-4f6e-9a49-0d5a0b1e6c11","type":"Unknown"}`))

	assert.Equal(t, []string{"h1"}, client.deletedHandles())
}

func TestHandle_UndecodableBodyIsDeleted(t *testing.T) {
	client := &fakeSQS{}
	c := newConsumer(client, map[string]Handler{"ReleaseStock": func(context.Context, Envelope) error {
		t.Fatal("handler must not run for an undecodable body")
		return nil
	}})

	c.handle(context.Background(), message("h1", "not-json"))

	assert.Equal(t, []string{"h1"}, client.deletedHandles())
}

func TestHandle_HandlerErrorKeepsMessageInQueue(t *testing.T) {
	client := &fakeSQS{}
	c := newConsumer(client, map[string]Handler{"ReleaseStock": func(context.Context, Envelope) error {
		return errors.New("database down")
	}})

	c.handle(context.Background(), message("h1", releaseBody))

	assert.Empty(t, client.deletedHandles(), "a failed message must be redelivered, not deleted")
}

type failingDeleteSQS struct{ fakeSQS }

func (f *failingDeleteSQS) DeleteMessage(context.Context, *sqs.DeleteMessageInput, ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	return nil, errors.New("throttled")
}

func TestHandle_DeleteFailureDoesNotPanic(t *testing.T) {
	c := newConsumer(&failingDeleteSQS{}, map[string]Handler{"ReleaseStock": func(context.Context, Envelope) error { return nil }})
	assert.NotPanics(t, func() { c.handle(context.Background(), message("h1", releaseBody)) })
}

func TestRun_ReceiveErrorIsRetriedUntilHandled(t *testing.T) {
	old := retryBackoff
	retryBackoff = time.Millisecond
	t.Cleanup(func() { retryBackoff = old })

	client := &fakeSQS{receiveEr: errors.New("network"), batches: [][]types.Message{{message("h1", releaseBody)}}}
	handled := make(chan struct{}, 1)
	c := newConsumer(client, map[string]Handler{"ReleaseStock": func(context.Context, Envelope) error {
		select {
		case handled <- struct{}{}:
		default:
		}
		return nil
	}})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		c.Run(ctx)
		close(done)
	}()

	select {
	case <-handled:
	case <-time.After(5 * time.Second):
		t.Fatal("consumer did not recover after a receive error")
	}
	cancel()
	<-done
}

func TestRun_StopsWhenContextIsCancelled(t *testing.T) {
	client := &errorAfterCancelSQS{}
	c := newConsumer(client, map[string]Handler{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c.Run(ctx)
	assert.False(t, client.called)
}

type errorAfterCancelSQS struct {
	fakeSQS
	called bool
}

func (f *errorAfterCancelSQS) ReceiveMessage(context.Context, *sqs.ReceiveMessageInput, ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	f.called = true
	return nil, errors.New("should not be called")
}

func TestStartAndStop(t *testing.T) {
	c := newConsumer(&fakeSQS{}, map[string]Handler{})
	assert.Equal(t, "sqs-consumer", c.Name())

	done := make(chan error, 1)
	go func() { done <- c.Start(context.Background()) }()

	require.NoError(t, c.Stop(context.Background()))
	require.NoError(t, c.Stop(context.Background())) // idempotent
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after Stop")
	}
}

func TestProcess_HandlerErrorIsReturned(t *testing.T) {
	c := newConsumer(&fakeSQS{}, map[string]Handler{"ReleaseStock": func(context.Context, Envelope) error {
		return errors.New("boom")
	}})
	ack, err := c.process(context.Background(), releaseBody)
	assert.False(t, ack)
	assert.Error(t, err)
}

func TestMemoryDedupe(t *testing.T) {
	d := NewMemoryDedupe()
	first, err := d.MarkProcessed(context.Background(), "id")
	require.NoError(t, err)
	second, err := d.MarkProcessed(context.Background(), "id")
	require.NoError(t, err)
	assert.True(t, first)
	assert.False(t, second)
}
