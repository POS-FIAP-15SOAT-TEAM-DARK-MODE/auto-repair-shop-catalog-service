package web

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
	jsonv2 "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/json"
	reservationDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	workDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
	"github.com/stretchr/testify/assert"
)

func TestError_StatusMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"unauthorized", auth.ErrInvalidCredentials, http.StatusUnauthorized},
		{"bad request work name", workDomain.ErrEmptyWorkName, http.StatusBadRequest},
		{"bad request supply id", supplyDomain.ErrInvalidSupplyID, http.StatusBadRequest},
		{"bad request supply id (alias)", supplyDomain.ErrInvalidSupplyId, http.StatusBadRequest},
		{"bad request supply name", supplyDomain.ErrInvalidSupplyName, http.StatusBadRequest},
		{"bad request supply description", supplyDomain.ErrInvalidSupplyDescription, http.StatusBadRequest},
		{"bad request supply unit price", supplyDomain.ErrInvalidSupplyUnitPrice, http.StatusBadRequest},
		{"bad request supply stock quantity", supplyDomain.ErrInvalidSupplyStockQuantity, http.StatusBadRequest},
		{"bad request supply version", supplyDomain.ErrInvalidSupplyVersion, http.StatusBadRequest},
		{"bad request supply version required", supplyDomain.ErrSupplyVersionRequired, http.StatusBadRequest},
		{"bad request supply amount", supplyDomain.ErrInvalidSupplyAmount, http.StatusBadRequest},
		{"bad request service order", reservationDomain.ErrInvalidServiceOrderID, http.StatusBadRequest},
		{"bad request json", jsonv2.ErrJSONSyntax, http.StatusBadRequest},
		{"joined errors are bad request", errors.Join(workDomain.ErrEmptyWorkName, workDomain.ErrWorkNameShorterThenRequired), http.StatusBadRequest},
		{"not found work", workDomain.ErrWorkNotFound, http.StatusNotFound},
		{"not found supply", supplyDomain.ErrSupplyNotFound, http.StatusNotFound},
		{"not found reservation", reservationDomain.ErrReservationNotFound, http.StatusNotFound},
		{"conflict already reserved", reservationDomain.ErrReservationAlreadyExists, http.StatusConflict},
		{"conflict insufficient stock", reservationDomain.ErrInsufficientStock, http.StatusConflict},
		{"conflict data", app.ErrDataConflict, http.StatusConflict},
		{"unprocessable quantity", reservationDomain.ErrInvalidReservationQuantity, http.StatusUnprocessableEntity},
		{"unprocessable out of stock", supplyDomain.ErrSupplyOutOfStock, http.StatusUnprocessableEntity},
		{"unprocessable data violation", app.ErrDataViolation, http.StatusUnprocessableEntity},
		{"internal json unsupported", jsonv2.ErrJSONUnsupportedValue, http.StatusInternalServerError},
		{"unknown error", fmt.Errorf("boom"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := Error(tt.err)
			assert.Equal(t, tt.want, status)
			assert.Equal(t, status, body.StatusCode)
		})
	}
}

func TestError_BodyShapes(t *testing.T) {
	_, body := Error(workDomain.ErrWorkNotFound)
	assert.Equal(t, workDomain.ErrWorkNotFound.Error(), body.Error)
	assert.Nil(t, body.Errors)

	_, body = Error(errors.Join(workDomain.ErrEmptyWorkName, workDomain.ErrWorkNameShorterThenRequired))
	assert.Len(t, body.Errors, 2)

	_, body = Error(fmt.Errorf("outer: %w", reservationDomain.ErrInsufficientStock))
	assert.Equal(t, reservationDomain.ErrInsufficientStock.Error(), body.Error, "domain 409 keeps its own message, without wrap context")

	_, body = Error(fmt.Errorf("x"))
	assert.Equal(t, "internal server error", body.Error, "500 never exposes the cause")
}

// pgLike mimics a driver error text that must never reach a client.
var pgLike = errors.New(`pq: duplicate key value violates unique constraint "supplies_pkey" (value 42)`)

func TestError_DoesNotLeakInfrastructureCause(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"500 unknown driver error", pgLike, http.StatusInternalServerError, "internal server error"},
		{"500 wrapped driver error", fmt.Errorf("repo: %w", pgLike), http.StatusInternalServerError, "internal server error"},
		{"500 json unsupported with cause", fmt.Errorf("%w: %v", jsonv2.ErrJSONUnsupportedValue, pgLike), http.StatusInternalServerError, "internal server error"},
		{"409 data conflict wrapping driver error", fmt.Errorf("%w: %v", app.ErrDataConflict, pgLike), http.StatusConflict, "conflict"},
		{"409 infra conflict", fmt.Errorf("save: %w", app.ErrInfraConflict), http.StatusConflict, "conflict"},
		{"422 data violation wrapping driver error", fmt.Errorf("%w: %v", app.ErrDataViolation, pgLike), http.StatusUnprocessableEntity, "invalid data"},
		{"409 domain keeps message", fmt.Errorf("%w: %v", reservationDomain.ErrReservationAlreadyExists, pgLike), http.StatusConflict, reservationDomain.ErrReservationAlreadyExists.Error()},
		{"422 domain keeps message", supplyDomain.ErrSupplyOutOfStock, http.StatusUnprocessableEntity, supplyDomain.ErrSupplyOutOfStock.Error()},
		{"404 domain keeps message", fmt.Errorf("find: %w", supplyDomain.ErrSupplyNotFound), http.StatusNotFound, supplyDomain.ErrSupplyNotFound.Error()},
		{"409 supply version conflict keeps message", supplyDomain.ErrSupplyVersionConflict, http.StatusConflict, supplyDomain.ErrSupplyVersionConflict.Error()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := Error(tt.err)
			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantStatus, body.StatusCode)
			assert.Equal(t, tt.wantMsg, body.Error)
			assert.NotContains(t, body.Error, "pq:")
			assert.NotContains(t, body.Error, "supplies_pkey")
		})
	}
}

func TestIsInfrastructureError(t *testing.T) {
	assert.True(t, IsInfrastructureError(fmt.Errorf("x: %w", app.ErrDataConflict)))
	assert.True(t, IsInfrastructureError(app.ErrInfraConflict))
	assert.True(t, IsInfrastructureError(app.ErrDataViolation))
	assert.False(t, IsInfrastructureError(pgLike))
	assert.False(t, IsInfrastructureError(supplyDomain.ErrSupplyOutOfStock))
}
