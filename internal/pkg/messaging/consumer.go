package messaging

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"go.uber.org/zap"
)

const (
	longPollSeconds = 20
	maxMessages     = 10
)

// retryBackoff is the pause after a failed ReceiveMessage call (a variable so tests can shorten it).
var retryBackoff = time.Second

// Handler applies one message type. It runs inside the unit of work that also
// records the message id, so a failed handler leaves no dedupe record behind.
type Handler func(ctx context.Context, env Envelope) error

// Consumer receives messages from one SQS queue and dispatches them by envelope type.
//
// Per message: dedupe + handler in one transaction, then DeleteMessage.
//   - handler error  -> message is not deleted (SQS redelivers, DLQ after maxReceiveCount)
//   - duplicate      -> deleted without running the handler again
//   - unknown type   -> logged and deleted
//   - undecodable    -> logged and deleted
type Consumer struct {
	queueURL string
	sqs      SQSClient
	uow      uow.Executor
	dedupe   Dedupe
	handlers map[string]Handler

	stopCh   chan struct{}
	stopOnce sync.Once
}

func NewConsumer(queueURL string, client SQSClient, unit uow.Executor, dedupe Dedupe, handlers map[string]Handler) *Consumer {
	return &Consumer{
		queueURL: queueURL,
		sqs:      client,
		uow:      unit,
		dedupe:   dedupe,
		handlers: handlers,
		stopCh:   make(chan struct{}),
	}
}

// Name implements app.Server.
func (c *Consumer) Name() string { return "sqs-consumer" }

// Start implements app.Server: it blocks until Stop is called or ctx is done.
func (c *Consumer) Start(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-c.stopCh:
			cancel()
		case <-runCtx.Done():
		}
	}()
	c.Run(runCtx)
	return nil
}

// Stop implements app.Server.
func (c *Consumer) Stop(_ context.Context) error {
	c.stopOnce.Do(func() { close(c.stopCh) })
	return nil
}

// Run polls the queue until ctx is done.
func (c *Consumer) Run(ctx context.Context) {
	for ctx.Err() == nil {
		out, err := c.sqs.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: maxMessages,
			WaitTimeSeconds:     longPollSeconds,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Global().Error(err, zap.String("operation", "sqs_receive"))
			select {
			case <-ctx.Done():
				return
			case <-time.After(retryBackoff):
			}
			continue
		}
		for _, m := range out.Messages {
			c.handle(ctx, m)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, m types.Message) {
	ack, err := c.process(ctx, aws.ToString(m.Body))
	if err != nil {
		logger.Global().Error(err, zap.String("operation", "sqs_process"))
		return
	}
	if !ack {
		return
	}
	if _, err := c.sqs.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: m.ReceiptHandle,
	}); err != nil {
		// The message will be redelivered; the dedupe record makes that harmless.
		logger.Global().Error(err, zap.String("operation", "sqs_delete"))
	}
}

// process returns ack=true when the message must be deleted from the queue.
func (c *Consumer) process(ctx context.Context, body string) (bool, error) {
	env, err := Decode([]byte(body))
	if err != nil {
		logger.Global().Warn("discarding undecodable message", zap.String("operation", "sqs_decode"))
		return true, nil
	}

	handler, ok := c.handlers[env.Type]
	if !ok {
		logger.Global().Info("discarding message of unknown type", zap.String("type", env.Type))
		return true, nil
	}

	err = c.uow.Execute(ctx, func(txCtx context.Context) error {
		first, err := c.dedupe.MarkProcessed(txCtx, env.MessageID)
		if err != nil {
			return err
		}
		if !first {
			logger.Global().Info("duplicate message ignored", zap.String("type", env.Type))
			return nil
		}
		return handler(txCtx, env)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
