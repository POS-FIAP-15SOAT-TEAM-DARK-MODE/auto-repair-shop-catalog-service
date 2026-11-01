//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// supplyView reads one supply from GET /v1/supplies, the way a client does before a PUT.
func (te *testEnv) supplyView(t *testing.T, supplyID string) (stock int, version int) {
	t.Helper()
	status, body := te.call(t, http.MethodGet, "/v1/supplies?page=1&pageSize=100000", "",
		map[string]string{"Authorization": "Bearer " + token(t, auth.ATTENDANT)})
	require.Equal(t, http.StatusOK, status, body)

	var page struct {
		Items []struct {
			ID            string `json:"id"`
			StockQuantity int    `json:"stock_quantity"`
			Version       int    `json:"version"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &page))
	for _, it := range page.Items {
		if it.ID == supplyID {
			return it.StockQuantity, it.Version
		}
	}
	t.Fatalf("supply %s not found in GET /v1/supplies", supplyID)
	return 0, 0
}

func putSupplyBody(stock, version int) string {
	return fmt.Sprintf(`{"name":"Filtro revisado","description":"Filtro de teste de integração","unitPrice":"35.00","stockQuantity":%d,"version":%d}`, stock, version)
}

// Reproduces the review scenario: a reservation between the client's read and its PUT
// must not be overwritten by the stale stock value.
func TestSupplyPutDoesNotOverwriteConcurrentReservation(t *testing.T) {
	te := setup(t)
	internal := map[string]string{"X-Internal-Token": internalToken}
	authz := map[string]string{"Authorization": "Bearer " + token(t, auth.ATTENDANT)}
	supplyID := te.insertSupply(t, 10)

	// 1. The client reads the supply: stock 10, version V.
	stockRead, versionRead := te.supplyView(t, supplyID)
	require.Equal(t, 10, stockRead)

	// 2. A reservation of 4 lands before the client's PUT.
	soID := "03" + supplyID[:24]
	status, body := te.call(t, http.MethodPost, "/internal/reservations", reserveBody(soID, supplyID, 4), internal)
	require.Equal(t, http.StatusCreated, status, body)
	require.Equal(t, 6, te.stock(t, supplyID))
	stockAfter, versionAfter := te.supplyView(t, supplyID)
	require.Equal(t, 6, stockAfter)
	require.Equal(t, versionRead+1, versionAfter)

	// 3. The stale PUT (stockQuantity=10, version V) must be rejected and keep the reservation.
	status, body = te.call(t, http.MethodPut, "/v1/supplies/"+supplyID, putSupplyBody(stockRead, versionRead), authz)
	assert.Equal(t, http.StatusConflict, status, body)
	assert.Contains(t, body, "supply was modified by someone else; reload and try again")
	assert.Equal(t, 6, te.stock(t, supplyID), "the reservation is not lost")

	// 4. A fresh read and a PUT with the current version succeed and bump the version again.
	stockNow, versionNow := te.supplyView(t, supplyID)
	require.Equal(t, 6, stockNow)
	status, body = te.call(t, http.MethodPut, "/v1/supplies/"+supplyID, putSupplyBody(stockNow, versionNow), authz)
	require.Equal(t, http.StatusOK, status, body)
	assert.Contains(t, body, fmt.Sprintf(`"version":%d`, versionNow+1))
	assert.Equal(t, 6, te.stock(t, supplyID))
}
