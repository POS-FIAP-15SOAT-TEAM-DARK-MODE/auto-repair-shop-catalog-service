//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// releaseStockCommand is a saga command as published by the Service Order service (messaging.md).
func releaseStockCommand(messageID, serviceOrderID string) string {
	return fmt.Sprintf(`{"message_id":%q,"saga_id":%q,"type":"ReleaseStock","occurred_at":"2026-11-01T12:00:00Z","service_order_id":%q,"request_id":%q,"payload":{}}`,
		messageID, uuid.NewString(), serviceOrderID, uuid.NewString())
}

func reserveBody(serviceOrderID, supplyID string, qty int) string {
	return fmt.Sprintf(`{"service_order_id":%q,"supply_id":%q,"quantity":%d}`, serviceOrderID, supplyID, qty)
}

func TestReservationFlow(t *testing.T) {
	te := setup(t)
	internal := map[string]string{"X-Internal-Token": internalToken}
	soID := "01" + uuid.NewString()[:24] // ULID-like, 26 chars
	supplyID := te.insertSupply(t, 10)

	status, body := te.call(t, http.MethodPost, "/internal/reservations", reserveBody(soID, supplyID, 2), internal)
	require.Equal(t, http.StatusCreated, status, body)
	assert.Equal(t, 8, te.stock(t, supplyID), "reserving decrements stock")

	status, _ = te.call(t, http.MethodPost, "/internal/reservations", reserveBody(uuid.NewString()[:26], supplyID, 9), internal)
	assert.Equal(t, http.StatusConflict, status, "insufficient stock")
	assert.Equal(t, 8, te.stock(t, supplyID))

	status, _ = te.call(t, http.MethodPost, "/internal/reservations", reserveBody(soID, supplyID, 1), internal)
	assert.Equal(t, http.StatusConflict, status, "second active reservation for the same order and supply")

	status, _ = te.call(t, http.MethodPost, "/internal/reservations", reserveBody(soID, uuid.NewString(), 1), internal)
	assert.Equal(t, http.StatusNotFound, status, "unknown supply")

	status, _ = te.call(t, http.MethodPost, "/internal/reservations", reserveBody(soID, supplyID, 0), internal)
	assert.Equal(t, http.StatusUnprocessableEntity, status, "quantity must be positive")

	status, _ = te.call(t, http.MethodPost, "/internal/reservations", reserveBody(soID, supplyID, 1), map[string]string{"X-Internal-Token": "wrong"})
	assert.Equal(t, http.StatusUnauthorized, status, "wrong internal token")
}

func TestReleaseStockDuplicateRestoresStockOnce(t *testing.T) {
	te := setup(t)
	internal := map[string]string{"X-Internal-Token": internalToken}
	soID := "02" + uuid.NewString()[:24]
	supplyID := te.insertSupply(t, 10)

	status, body := te.call(t, http.MethodPost, "/internal/reservations", reserveBody(soID, supplyID, 2), internal)
	require.Equal(t, http.StatusCreated, status, body)
	require.Equal(t, 8, te.stock(t, supplyID))

	// The same command is delivered twice (at-least-once delivery).
	msg := releaseStockCommand(uuid.NewString(), soID)
	te.send(t, msg)
	te.send(t, msg)

	require.Eventually(t, func() bool { return te.stock(t, supplyID) == 10 }, 60*time.Second, 500*time.Millisecond,
		"ReleaseStock must restore the reserved units")
	time.Sleep(3 * time.Second) // give a duplicate delivery time to be processed
	assert.Equal(t, 10, te.stock(t, supplyID), "stock is restored exactly once")
}

func TestPublicAPIRequiresJWT(t *testing.T) {
	te := setup(t)

	status, _ := te.call(t, http.MethodGet, "/v1/works", "", nil)
	assert.Equal(t, http.StatusUnauthorized, status)

	status, body := te.call(t, http.MethodGet, "/v1/works", "", map[string]string{
		"Authorization": "Bearer " + token(t, auth.ATTENDANT),
	})
	assert.Equal(t, http.StatusOK, status, body)
}

func TestInternalWorkLookup(t *testing.T) {
	te := setup(t)
	id := uuid.NewString()
	_, err := te.db.Exec(`INSERT INTO "work" (id, name, description, unit_price, status) VALUES ($1, $2, $3, $4, $5)`,
		id, "Troca de óleo", "Troca de óleo do motor", "150.00", true)
	require.NoError(t, err)

	status, body := te.call(t, http.MethodGet, "/internal/works/"+id, "", map[string]string{"X-Internal-Token": internalToken})
	require.Equal(t, http.StatusOK, status, body)
	assert.Contains(t, body, `"unit_price":"150.00"`)
	assert.Contains(t, body, `"active":true`)

	status, _ = te.call(t, http.MethodGet, "/internal/works/"+uuid.NewString(), "", map[string]string{"X-Internal-Token": internalToken})
	assert.Equal(t, http.StatusNotFound, status)
}
