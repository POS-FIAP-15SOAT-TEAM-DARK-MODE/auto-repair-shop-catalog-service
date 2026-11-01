package domain

import (
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Status of a stock reservation. A reservation is RESERVED while the supply
// units are held for a service order and RELEASED once they are returned.
type Status string

const (
	StatusReserved Status = "RESERVED"
	StatusReleased Status = "RELEASED"
)

type Reservation struct {
	ID             string
	ServiceOrderID string
	SupplyID       string
	Quantity       int
	Status         Status
}

// SupplyInfo is the snapshot of a supply the reservation needs to answer the caller.
type SupplyInfo struct {
	ID            string
	Name          string
	UnitPrice     decimal.Decimal
	StockQuantity int
}

// NewReservation validates the input and builds a RESERVED reservation with a new ID.
func NewReservation(serviceOrderID, supplyID string, quantity int) (*Reservation, error) {
	if strings.TrimSpace(serviceOrderID) == "" {
		return nil, ErrInvalidServiceOrderID
	}
	if strings.TrimSpace(supplyID) == "" {
		return nil, domain.ErrInvalidSupplyID
	}
	if quantity <= 0 {
		return nil, ErrInvalidReservationQuantity
	}

	return &Reservation{
		ID:             uuid.NewString(),
		ServiceOrderID: serviceOrderID,
		SupplyID:       supplyID,
		Quantity:       quantity,
		Status:         StatusReserved,
	}, nil
}
