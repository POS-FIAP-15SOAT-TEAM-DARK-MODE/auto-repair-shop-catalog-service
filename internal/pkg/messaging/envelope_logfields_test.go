package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func fieldKeys(fields []zap.Field) map[string]string {
	out := map[string]string{}
	for _, f := range fields {
		out[f.Key] = f.String
	}
	return out
}

func TestEnvelope_LogFields_IncludesPresentIdentifiers(t *testing.T) {
	env := Envelope{
		MessageID:      "b7a1c6f0-2f64-4f6e-9a49-0d5a0b1e6c11",
		SagaID:         "saga-1",
		Type:           "ReleaseStock",
		ServiceOrderID: "so-1",
		RequestID:      "req-1",
	}

	got := fieldKeys(env.LogFields())

	assert.Equal(t, env.MessageID, got["message_id"])
	assert.Equal(t, "ReleaseStock", got["message_type"])
	assert.Equal(t, "saga-1", got["saga_id"])
	assert.Equal(t, "so-1", got["service_order_id"])
	assert.Equal(t, "req-1", got["request_id"])
}

func TestEnvelope_LogFields_OmitsEmptyIdentifiers(t *testing.T) {
	env := Envelope{MessageID: "b7a1c6f0-2f64-4f6e-9a49-0d5a0b1e6c11", Type: "ReleaseStock"}

	got := fieldKeys(env.LogFields())

	assert.Len(t, env.LogFields(), 2)
	assert.NotContains(t, got, "saga_id")
	assert.NotContains(t, got, "service_order_id")
	assert.NotContains(t, got, "request_id")
}
