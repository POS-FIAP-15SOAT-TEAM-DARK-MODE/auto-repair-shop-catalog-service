package routing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
	reservationAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/adapters"
	supplyAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	workAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const internalToken = "internal-secret"

type okWork struct{}

func (okWork) Create(context.Context, *http.Request) (workAdapters.WorkResponse, error) {
	return workAdapters.WorkResponse{}, nil
}
func (okWork) List(context.Context, *http.Request) (workAdapters.PaginatedWorkResponse, error) {
	return workAdapters.PaginatedWorkResponse{}, nil
}
func (okWork) Update(context.Context, *http.Request) (workAdapters.WorkResponse, error) {
	return workAdapters.WorkResponse{}, nil
}
func (okWork) Delete(context.Context, *http.Request) error { return nil }

type okWorkInternal struct{}

func (okWorkInternal) Get(context.Context, *http.Request) (workAdapters.InternalWorkResponse, error) {
	return workAdapters.InternalWorkResponse{ID: "w1", UnitPrice: "150.00", Active: true}, nil
}

type okSupply struct{}

func (okSupply) Create(context.Context, *http.Request) (supplyAdapters.SupplyResponse, error) {
	return supplyAdapters.SupplyResponse{}, nil
}
func (okSupply) List(context.Context, *http.Request) (supplyAdapters.PaginatedSupplyResponse, error) {
	return supplyAdapters.PaginatedSupplyResponse{}, nil
}
func (okSupply) Update(context.Context, *http.Request) (supplyAdapters.SupplyResponse, error) {
	return supplyAdapters.SupplyResponse{}, nil
}
func (okSupply) Delete(context.Context, *http.Request) error { return nil }

type okReservation struct{}

func (okReservation) Reserve(context.Context, *http.Request) (reservationAdapters.ReserveResponse, error) {
	return reservationAdapters.ReserveResponse{ReservationID: "r1"}, nil
}
func (okReservation) Release(context.Context, *http.Request) error { return nil }

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	c := &container.HandlersWrapper{
		WorkHandler:         okWork{},
		WorkInternalHandler: okWorkInternal{},
		SupplyHandler:       okSupply{},
		ReservationHandler:  okReservation{},
	}
	return SetupRouter(c, &container.Middlewares{}, internalToken)
}

func do(router http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestRouting_PublicV1RequiresJWT(t *testing.T) {
	router := newRouter()
	assert.Equal(t, http.StatusOK, do(router, http.MethodGet, "/ping", nil).Code)
	assert.Equal(t, http.StatusUnauthorized, do(router, http.MethodGet, "/v1/works", nil).Code)

	token, err := auth.GenerateToken("u", []auth.Role{auth.ATTENDANT}, timeNowPlusHour())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, do(router, http.MethodGet, "/v1/works", map[string]string{"Authorization": "Bearer " + token}).Code)
	assert.Equal(t, http.StatusOK, do(router, http.MethodGet, "/v1/supplies", map[string]string{"Authorization": "Bearer " + token}).Code)
	assert.Equal(t, http.StatusForbidden, do(router, http.MethodDelete, "/v1/works/1", map[string]string{"Authorization": "Bearer " + tokenFor(t, auth.MECHANIC)}).Code)
}

func TestRouting_InternalRequiresToken(t *testing.T) {
	router := newRouter()
	assert.Equal(t, http.StatusUnauthorized, do(router, http.MethodGet, "/internal/works/w1", nil).Code)
	assert.Equal(t, http.StatusUnauthorized, do(router, http.MethodPost, "/internal/reservations", map[string]string{"X-Internal-Token": "bad"}).Code)

	ok := map[string]string{"X-Internal-Token": internalToken}
	assert.Equal(t, http.StatusOK, do(router, http.MethodGet, "/internal/works/w1", ok).Code)
	assert.Equal(t, http.StatusCreated, do(router, http.MethodPost, "/internal/reservations", ok).Code)
	assert.Equal(t, http.StatusNoContent, do(router, http.MethodDelete, "/internal/reservations/r1", ok).Code)
}

func TestRouting_InternalIsNotPublic(t *testing.T) {
	router := newRouter()
	assert.Equal(t, http.StatusNotFound, do(router, http.MethodGet, "/v1/internal/works/w1", nil).Code)
}
