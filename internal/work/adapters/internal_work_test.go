package adapters

import (
	"encoding/json"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The JSON must match contracts/internal-rest.md (GET /internal/works/{id}).
func TestInternalWorkResponse_MatchesContract(t *testing.T) {
	resp := InternalWorkDomainToResponse(domain.Work{
		ID:          "9b2d6a0e-3c0f-4a0b-8d57-2a5f6a1c9e11",
		Name:        "Troca de óleo",
		Description: "Troca de óleo do motor com filtro incluído",
		Price:       decimal.RequireFromString("150"),
		Status:      domain.ACTIVE,
	})

	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":"9b2d6a0e-3c0f-4a0b-8d57-2a5f6a1c9e11","name":"Troca de óleo","description":"Troca de óleo do motor com filtro incluído","unit_price":"150.00","active":true}`, string(raw))
}
