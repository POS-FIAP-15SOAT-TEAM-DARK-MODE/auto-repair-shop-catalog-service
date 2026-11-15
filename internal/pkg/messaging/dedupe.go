package messaging

import (
	"context"
	"sync"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
)

// Dedupe records processed message ids. MarkProcessed must run inside the unit
// of work carried by ctx so the record commits together with the handler effect.
type Dedupe interface {
	// MarkProcessed returns true when the id was not seen before.
	MarkProcessed(ctx context.Context, messageID string) (bool, error)
}

type postgresDedupe struct{}

const markProcessedQuery = `INSERT INTO processed_message (message_id) VALUES ($1) ON CONFLICT DO NOTHING`

func NewPostgresDedupe() Dedupe {
	return &postgresDedupe{}
}

func (postgresDedupe) MarkProcessed(ctx context.Context, messageID string) (bool, error) {
	tx, err := uow.GetTransaction(ctx)
	if err != nil {
		return false, err
	}
	res, err := tx.ExecContext(ctx, markProcessedQuery, messageID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// NewMemoryDedupe is an in-memory Dedupe for tests.
func NewMemoryDedupe() Dedupe {
	return &memoryDedupe{seen: map[string]struct{}{}}
}

type memoryDedupe struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

func (m *memoryDedupe) MarkProcessed(_ context.Context, messageID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.seen[messageID]; ok {
		return false, nil
	}
	m.seen[messageID] = struct{}{}
	return true, nil
}
