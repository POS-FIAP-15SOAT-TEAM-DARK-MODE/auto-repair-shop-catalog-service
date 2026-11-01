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
	assert.Equal(t, reservationDomain.ErrInsufficientStock.Error(), body.Error, "409/422/5xx expose the root cause")

	_, body = Error(fmt.Errorf("x"))
	assert.Equal(t, "x", body.Error, "500 exposes the root cause message")
}
