package reservation

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/adapters"
	"github.com/stretchr/testify/assert"
)

type recordingService struct {
	reserveReq adapters.ReserveRequest
	released   string
	err        error
}

func (s *recordingService) Reserve(_ context.Context, req adapters.ReserveRequest) (adapters.ReserveResponse, error) {
	s.reserveReq = req
	return adapters.ReserveResponse{ReservationID: "r1"}, s.err
}

func (s *recordingService) Release(_ context.Context, id string) error {
	s.released = id
	return s.err
}

func (s *recordingService) ReleaseAllByServiceOrder(context.Context, string) error { return nil }

func TestController_ReserveDecodesBody(t *testing.T) {
	svc := &recordingService{}
	ctrl := NewController(svc)
	req := httptest.NewRequest(http.MethodPost, "/internal/reservations",
		strings.NewReader(`{"service_order_id":"so","supply_id":"s1","quantity":2}`))

	resp, err := ctrl.Reserve(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, "r1", resp.ReservationID)
	assert.Equal(t, adapters.ReserveRequest{ServiceOrderID: "so", SupplyID: "s1", Quantity: 2}, svc.reserveReq)
}

func TestController_ReserveRejectsBadJSON(t *testing.T) {
	ctrl := NewController(&recordingService{})
	req := httptest.NewRequest(http.MethodPost, "/internal/reservations", strings.NewReader(`{`))
	_, err := ctrl.Reserve(context.Background(), req)
	assert.Error(t, err)
}

func TestController_ReleaseReadsIDFromContext(t *testing.T) {
	svc := &recordingService{err: errors.New("x")}
	ctrl := NewController(svc)
	ctx := context.WithValue(context.Background(), idParam, "r9") //nolint:staticcheck // route params are string keys by design

	err := ctrl.Release(ctx, nil)

	assert.Error(t, err)
	assert.Equal(t, "r9", svc.released)
}
