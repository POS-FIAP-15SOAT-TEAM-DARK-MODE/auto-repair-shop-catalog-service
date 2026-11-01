package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidEnvelope is returned when a message body is not a valid envelope (messaging.md).
var ErrInvalidEnvelope = errors.New("invalid message envelope")

// Envelope is the body of every message exchanged through SNS/SQS.
type Envelope struct {
	MessageID      string          `json:"message_id"`
	SagaID         string          `json:"saga_id"`
	Type           string          `json:"type"`
	OccurredAt     time.Time       `json:"occurred_at"`
	ServiceOrderID string          `json:"service_order_id"`
	RequestID      string          `json:"request_id"`
	Payload        json.RawMessage `json:"payload"`
}

// Decode parses and validates a message body. message_id must be a UUID because
// it is the deduplication key; type must be present.
func Decode(body []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return Envelope{}, fmt.Errorf("%w: %s", ErrInvalidEnvelope, err.Error())
	}
	if _, err := uuid.Parse(env.MessageID); err != nil {
		return Envelope{}, fmt.Errorf("%w: message_id", ErrInvalidEnvelope)
	}
	if env.Type == "" {
		return Envelope{}, fmt.Errorf("%w: type", ErrInvalidEnvelope)
	}
	return env, nil
}
