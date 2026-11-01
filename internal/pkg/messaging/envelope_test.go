package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecode_ValidEnvelope(t *testing.T) {
	body := []byte(`{"message_id":"b7a1c6f0-2f64-4f6e-9a49-0d5a0b1e6c11","saga_id":"0f3c9a2e-6a1d-4a53-8a6e-5d3b1f2c7e90","type":"ReleaseStock","occurred_at":"2026-11-01T12:00:00Z","service_order_id":"01JD8Z4R3Q8M2V7K5N9XWABCDE","request_id":"2c5e1f0a-9d77-4b2c-a0c1-3f8e6b7d1a22","payload":{}}`)

	env, err := Decode(body)

	require.NoError(t, err)
	assert.Equal(t, "ReleaseStock", env.Type)
	assert.Equal(t, "01JD8Z4R3Q8M2V7K5N9XWABCDE", env.ServiceOrderID)
	assert.Equal(t, "2026-11-01T12:00:00Z", env.OccurredAt.UTC().Format("2006-01-02T15:04:05Z"))
	assert.JSONEq(t, "{}", string(env.Payload))
}

func TestDecode_Rejects(t *testing.T) {
	tests := map[string]string{
		"not json":            `{"message_id":`,
		"message id not uuid": `{"message_id":"abc","type":"ReleaseStock"}`,
		"missing type":        `{"message_id":"b7a1c6f0-2f64-4f6e-9a49-0d5a0b1e6c11"}`,
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Decode([]byte(body))
			assert.ErrorIs(t, err, ErrInvalidEnvelope)
		})
	}
}
