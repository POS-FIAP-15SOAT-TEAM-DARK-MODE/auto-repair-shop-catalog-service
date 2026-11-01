package interfaces

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/domain"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ReservationService --with-expecter
type ReservationService interface {
	Reserve(ctx context.Context, req adapters.ReserveRequest) (adapters.ReserveResponse, error)
	Release(ctx context.Context, reservationID string) error
	ReleaseAllByServiceOrder(ctx context.Context, serviceOrderID string) error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=ReservationHTTPController --with-expecter
type ReservationHTTPController interface {
	Reserve(ctx context.Context, req *http.Request) (adapters.ReserveResponse, error)
	Release(ctx context.Context, req *http.Request) error
}

// ReservationRepository persists reservations. Writes run inside the unit of work
// carried by the context.
type ReservationRepository interface {
	// Reserve inserts a reservation, or reactivates a RELEASED one for the same
	// (service order, supply) pair. Returns ErrReservationAlreadyExists when an
	// active reservation already exists. On success r.ID holds the stored id.
	Reserve(ctx context.Context, r *domain.Reservation) error
	FindByID(ctx context.Context, id string) (domain.Reservation, error)
	// Release flips a RESERVED reservation to RELEASED and returns it. It returns
	// (nil, nil) when there is no active reservation with that id.
	Release(ctx context.Context, id string) (*domain.Reservation, error)
	// ReleaseAllByServiceOrder releases every RESERVED reservation of the service
	// order and returns the ones it released.
	ReleaseAllByServiceOrder(ctx context.Context, serviceOrderID string) ([]domain.Reservation, error)
}

// StockRepository reads and moves supply stock inside the unit of work.
type StockRepository interface {
	FindSupply(ctx context.Context, id string) (domain.SupplyInfo, error)
	// Decrement removes qty units; returns ErrInsufficientStock when stock is not enough.
	Decrement(ctx context.Context, supplyID string, qty int) error
	Increment(ctx context.Context, supplyID string, qty int) error
}
