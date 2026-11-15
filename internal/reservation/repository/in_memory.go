package repository

import (
	"context"
	"sync"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/shopspring/decimal"
)

// InMemory implements ReservationRepository and StockRepository for unit tests.
// It has no rollback: callers must validate before writing, as the service does.
type InMemory struct {
	mu           sync.Mutex
	reservations map[string]domain.Reservation
	supplies     map[string]domain.SupplyInfo
}

func NewInMemory() *InMemory {
	return &InMemory{
		reservations: map[string]domain.Reservation{},
		supplies:     map[string]domain.SupplyInfo{},
	}
}

// AddSupply seeds a supply with the given stock.
func (m *InMemory) AddSupply(id, name, description string, unitPrice decimal.Decimal, stock int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.supplies[id] = domain.SupplyInfo{ID: id, Name: name, Description: description, UnitPrice: unitPrice, StockQuantity: stock, Version: 1}
}

func (m *InMemory) Stock(id string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.supplies[id].StockQuantity
}

func (m *InMemory) Reserve(_ context.Context, r *domain.Reservation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, existing := range m.reservations {
		if existing.ServiceOrderID == r.ServiceOrderID && existing.SupplyID == r.SupplyID {
			if existing.Status == domain.StatusReserved {
				return domain.ErrReservationAlreadyExists
			}
			existing.Status = domain.StatusReserved
			existing.Quantity = r.Quantity
			m.reservations[id] = existing
			r.ID = id
			r.Status = domain.StatusReserved
			return nil
		}
	}
	m.reservations[r.ID] = *r
	return nil
}

func (m *InMemory) FindByID(_ context.Context, id string) (domain.Reservation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.reservations[id]
	if !ok {
		return domain.Reservation{}, domain.ErrReservationNotFound
	}
	return r, nil
}

func (m *InMemory) Release(_ context.Context, id string) (*domain.Reservation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.reservations[id]
	if !ok || r.Status != domain.StatusReserved {
		return nil, nil
	}
	r.Status = domain.StatusReleased
	m.reservations[id] = r
	return &r, nil
}

func (m *InMemory) ReleaseAllByServiceOrder(_ context.Context, serviceOrderID string) ([]domain.Reservation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var released []domain.Reservation
	for id, r := range m.reservations {
		if r.ServiceOrderID == serviceOrderID && r.Status == domain.StatusReserved {
			r.Status = domain.StatusReleased
			m.reservations[id] = r
			released = append(released, r)
		}
	}
	return released, nil
}

func (m *InMemory) FindSupply(_ context.Context, id string) (domain.SupplyInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.supplies[id]
	if !ok {
		return domain.SupplyInfo{}, supplyDomain.ErrSupplyNotFound
	}
	return s, nil
}

func (m *InMemory) Decrement(_ context.Context, supplyID string, qty int) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.supplies[supplyID]
	if !ok || s.StockQuantity < qty {
		return 0, domain.ErrInsufficientStock
	}
	s.StockQuantity -= qty
	s.Version++
	m.supplies[supplyID] = s
	return s.Version, nil
}

func (m *InMemory) Increment(_ context.Context, supplyID string, qty int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.supplies[supplyID]
	s.StockQuantity += qty
	m.supplies[supplyID] = s
	return nil
}
