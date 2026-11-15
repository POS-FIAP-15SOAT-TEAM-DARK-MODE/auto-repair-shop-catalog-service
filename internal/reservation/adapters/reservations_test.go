package adapters

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The 201 body must match contracts/internal-rest.md (POST /internal/reservations).
func TestReserveResponse_MatchesContract(t *testing.T) {
	raw, err := json.Marshal(ReserveResponse{
		ReservationID: "r-1",
		SupplyID:      "s-1",
		Name:          "Filtro",
		Description:   "Filtro de óleo",
		UnitPrice:     "35.50",
		Version:       3,
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{"reservation_id":"r-1","supply_id":"s-1","name":"Filtro","description":"Filtro de óleo","unit_price":"35.50","version":3}`, string(raw))
}
