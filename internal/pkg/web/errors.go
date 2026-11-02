package web

import (
	"errors"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/auth"
	jsonv2 "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/json"
	reservationDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	workDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
)

const (
	msgInternalServerError = "internal server error"
	msgConflict            = "conflict"
	msgInvalidData         = "invalid data"
)

// Domain sentinels per client-facing status. Their messages are safe to expose.
var (
	notFoundErrors = []error{
		workDomain.ErrWorkNotFound,
		supplyDomain.ErrSupplyNotFound,
		reservationDomain.ErrReservationNotFound,
	}
	domainConflictErrors = []error{
		reservationDomain.ErrReservationAlreadyExists,
		supplyDomain.ErrSupplyVersionConflict,
		reservationDomain.ErrInsufficientStock,
	}
	domainUnprocessableErrors = []error{
		workDomain.ErrWorkPriceLessThenOrEqualZero,
		workDomain.ErrInvalidWorkPriceValue,
		supplyDomain.ErrSupplyOutOfStock,
		reservationDomain.ErrInvalidReservationQuantity,
	}
	// infrastructureErrors come from the persistence layer. Their messages are
	// generic on purpose: the real cause is only logged.
	infrastructureErrors = []error{
		app.ErrDataConflict,
		app.ErrInfraConflict,
		app.ErrDataViolation,
	}
)

type errorResponse struct {
	StatusCode int      `json:"code"`
	Errors     []string `json:"errors,omitempty"`
	Error      string   `json:"error,omitempty"`
}

// Error builds an errorResponse for the given error and decides the status code.
func Error(err error) (int, errorResponse) {
	status := getHTTPStatus(err)
	stackedErrors, errorMsg := buildErrorMessages(err, status)

	return status, errorResponse{
		StatusCode: status,
		Errors:     stackedErrors,
		Error:      errorMsg,
	}
}

// IsInfrastructureError reports whether err stems from persistence/infrastructure.
// Such errors must be logged in full and never reach the response body.
func IsInfrastructureError(err error) bool {
	return matchesAny(err, infrastructureErrors)
}

// getHTTPStatus determines the HTTP status code for a given error.
func getHTTPStatus(err error) int {
	switch {
	case isUnauthorizedError(err):
		return http.StatusUnauthorized
	case isBadRequestError(err), IsJoined(err):
		return http.StatusBadRequest
	case isNotFoundError(err):
		return http.StatusNotFound
	case isInternalServerError(err):
		return http.StatusInternalServerError
	case isConflictError(err):
		return http.StatusConflict
	case isUnprocessableEntityError(err):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func IsJoined(err error) bool {
	type unwrapper interface {
		Unwrap() []error
	}

	var uw unwrapper
	return errors.As(err, &uw)
}

// buildErrorMessages formats the error messages based on status code.
// Infrastructure causes never leave the server: 5xx and infra 409/422 get generic messages.
func buildErrorMessages(err error, status int) ([]string, string) {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest:
		// For 4XX BadRequest provide stack of errors.
		return unwrapAll(err), ""
	case http.StatusNotFound:
		return nil, domainMessage(err, notFoundErrors)
	case http.StatusConflict:
		if IsInfrastructureError(err) {
			return nil, msgConflict
		}
		return nil, domainMessage(err, domainConflictErrors)
	case http.StatusUnprocessableEntity:
		if IsInfrastructureError(err) {
			return nil, msgInvalidData
		}
		return nil, domainMessage(err, domainUnprocessableErrors)
	default:
		return nil, msgInternalServerError
	}
}

// domainMessage returns the message of the domain sentinel matched by err.
// Wrapping context and infrastructure causes are never included.
func domainMessage(err error, sentinels []error) string {
	for _, s := range sentinels {
		if errors.Is(err, s) {
			return s.Error()
		}
	}
	return msgInternalServerError
}

func matchesAny(err error, sentinels []error) bool {
	for _, s := range sentinels {
		if errors.Is(err, s) {
			return true
		}
	}
	return false
}

func isBadRequestError(err error) bool {
	return errors.Is(err, workDomain.ErrEmptyWorkName) ||
		errors.Is(err, workDomain.ErrWorkNameShorterThenRequired) ||
		errors.Is(err, workDomain.ErrEmptyWorkDescription) ||
		errors.Is(err, workDomain.ErrWorkDescriptionShorterThenRequired) ||
		errors.Is(err, workDomain.ErrInvalidWorkId) ||
		errors.Is(err, workDomain.ErrInvalidWorkStatusValue) ||
		errors.Is(err, supplyDomain.ErrInvalidSupplyID) ||
		errors.Is(err, supplyDomain.ErrSupplyVersionRequired) ||
		errors.Is(err, supplyDomain.ErrInvalidSupplyAmount) ||
		errors.Is(err, reservationDomain.ErrInvalidServiceOrderID) ||
		errors.Is(err, jsonv2.ErrJSONSyntax) ||
		errors.Is(err, jsonv2.ErrJSONType) ||
		errors.Is(err, jsonv2.ErrJSONUnexpectedEOF) ||
		errors.Is(err, jsonv2.ErrJSONEmptyBody) ||
		errors.Is(err, jsonv2.ErrWrongPayloadFormat)
}

func isNotFoundError(err error) bool {
	return matchesAny(err, notFoundErrors)
}

func isUnauthorizedError(err error) bool {
	return errors.Is(err, auth.ErrInvalidCredentials)
}

func isInternalServerError(err error) bool {
	return errors.Is(err, jsonv2.ErrJSONUnsupportedValue) ||
		errors.Is(err, jsonv2.ErrJSONUnsupportedType) ||
		errors.Is(err, jsonv2.ErrJSONInvalidUnmarshal)
}

func isConflictError(err error) bool {
	return errors.Is(err, app.ErrDataConflict) ||
		errors.Is(err, app.ErrInfraConflict) ||
		matchesAny(err, domainConflictErrors)
}

func isUnprocessableEntityError(err error) bool {
	return errors.Is(err, app.ErrDataViolation) ||
		matchesAny(err, domainUnprocessableErrors)
}

func unwrapAll(err error) []string {
	if err == nil {
		return nil
	}

	type multiUnwrap interface {
		Unwrap() []error
	}
	if m, ok := err.(multiUnwrap); ok {
		var all []string
		for _, e := range m.Unwrap() {
			all = append(all, unwrapAll(e)...)
		}
		return all
	}
	return []string{err.Error()}
}
