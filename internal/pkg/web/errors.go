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

// Known sentinels grouped by the client-facing answer. Their messages are safe to expose.
var (
	unauthorizedErrors = []error{
		auth.ErrInvalidCredentials,
	}
	// validationErrors answer 400; their messages go to errors[].
	validationErrors = []error{
		workDomain.ErrEmptyWorkName,
		workDomain.ErrWorkNameShorterThenRequired,
		workDomain.ErrEmptyWorkDescription,
		workDomain.ErrWorkDescriptionShorterThenRequired,
		workDomain.ErrInvalidWorkId,
		workDomain.ErrInvalidWorkStatusValue,
		workDomain.ErrWorkPriceLessThenOrEqualZero,
		workDomain.ErrInvalidWorkPriceValue,
		supplyDomain.ErrInvalidSupplyID,
		supplyDomain.ErrInvalidSupplyId,
		supplyDomain.ErrInvalidSupplyName,
		supplyDomain.ErrInvalidSupplyDescription,
		supplyDomain.ErrInvalidSupplyUnitPrice,
		supplyDomain.ErrInvalidSupplyStockQuantity,
		supplyDomain.ErrInvalidSupplyVersion,
		supplyDomain.ErrSupplyVersionRequired,
		supplyDomain.ErrInvalidSupplyAmount,
		reservationDomain.ErrInvalidServiceOrderID,
		jsonv2.ErrJSONSyntax,
		jsonv2.ErrJSONType,
		jsonv2.ErrJSONUnexpectedEOF,
		jsonv2.ErrJSONEmptyBody,
		jsonv2.ErrWrongPayloadFormat,
	}
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
	// serializationErrors are JSON encoding problems on our side: 500.
	serializationErrors = []error{
		jsonv2.ErrJSONUnsupportedValue,
		jsonv2.ErrJSONUnsupportedType,
		jsonv2.ErrJSONInvalidUnmarshal,
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
// Any validation sentinel among its causes makes it a 400, whatever else it joins.
func getHTTPStatus(err error) int {
	switch {
	case isUnauthorizedError(err):
		return http.StatusUnauthorized
	case isValidationError(err):
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

// buildErrorMessages formats the error messages based on status code.
// Bodies only ever contain sentinel messages, never wrap context or driver text.
func buildErrorMessages(err error, status int) ([]string, string) {
	switch status {
	case http.StatusBadRequest:
		return sentinelMessages(err, validationErrors), ""
	case http.StatusUnauthorized:
		return sentinelMessages(err, unauthorizedErrors), ""
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

// domainMessage returns the message of the first domain sentinel matched by err.
// Wrapping context and infrastructure causes are never included.
func domainMessage(err error, sentinels []error) string {
	for _, s := range sentinels {
		if errors.Is(err, s) {
			return s.Error()
		}
	}
	return msgInternalServerError
}

// sentinelMessages returns, in order, the message of the sentinel matched by each
// cause of err that belongs to the given list. Other causes are never included.
func sentinelMessages(err error, sentinels []error) []string {
	var msgs []string
	for _, cause := range causes(err) {
		for _, s := range sentinels {
			if errors.Is(cause, s) {
				msgs = append(msgs, s.Error())
				break
			}
		}
	}
	return msgs
}

// causes returns the innermost errors of err, following every branch of joins and wraps in order.
func causes(err error) []error {
	if err == nil {
		return nil
	}

	type multiUnwrap interface {
		Unwrap() []error
	}
	type singleUnwrap interface {
		Unwrap() error
	}

	switch u := err.(type) {
	case multiUnwrap:
		var out []error
		for _, e := range u.Unwrap() {
			out = append(out, causes(e)...)
		}
		return out
	case singleUnwrap:
		if inner := u.Unwrap(); inner != nil {
			return causes(inner)
		}
	}
	return []error{err}
}

func matchesAny(err error, sentinels []error) bool {
	for _, s := range sentinels {
		if errors.Is(err, s) {
			return true
		}
	}
	return false
}

func isValidationError(err error) bool {
	return len(sentinelMessages(err, validationErrors)) > 0
}

func isNotFoundError(err error) bool {
	return matchesAny(err, notFoundErrors)
}

func isUnauthorizedError(err error) bool {
	return matchesAny(err, unauthorizedErrors)
}

func isInternalServerError(err error) bool {
	return matchesAny(err, serializationErrors)
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
