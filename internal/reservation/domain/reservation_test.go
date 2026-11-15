package domain

import (
	"testing"

	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReservation(t *testing.T) {
	r, err := NewReservation("01JD8Z4R3Q8M2V7K5N9XWABCDE", "supply-1", 2)
	require.NoError(t, err)
	assert.NotEmpty(t, r.ID)
	assert.Equal(t, StatusReserved, r.Status)
	assert.Equal(t, 2, r.Quantity)
}

func TestNewReservation_Validation(t *testing.T) {
	_, err := NewReservation("  ", "supply-1", 1)
	assert.ErrorIs(t, err, ErrInvalidServiceOrderID)

	_, err = NewReservation("so", " ", 1)
	assert.ErrorIs(t, err, supplyDomain.ErrInvalidSupplyID)

	_, err = NewReservation("so", "supply-1", 0)
	assert.ErrorIs(t, err, ErrInvalidReservationQuantity)
}
