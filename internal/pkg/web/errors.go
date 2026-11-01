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
func buildErrorMessages(err error, status int) ([]string, string) {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest:
		// For 4XX BadRequest provide stack of errors.
		return unwrapAll(err), ""
	case http.StatusNotFound:
		return nil, err.Error()
	case http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError:
		// For 409/422 and 5XX unwrap recursively and return the innermost (root cause) error message.
		root := err
		for {
			unwrapped := errors.Unwrap(root)
			if unwrapped == nil {
				break
			}
			root = unwrapped
		}
		return nil, root.Error()
	default:
		// For any other case, generic error message.
		return nil, "Internal Server Error"
	}
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
	return errors.Is(err, workDomain.ErrWorkNotFound) ||
		errors.Is(err, supplyDomain.ErrSupplyNotFound) ||
		errors.Is(err, reservationDomain.ErrReservationNotFound)
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
		errors.Is(err, reservationDomain.ErrReservationAlreadyExists) ||
		errors.Is(err, supplyDomain.ErrSupplyVersionConflict) ||
		errors.Is(err, reservationDomain.ErrInsufficientStock)
}

func isUnprocessableEntityError(err error) bool {
	return errors.Is(err, app.ErrDataViolation) ||
		errors.Is(err, workDomain.ErrWorkPriceLessThenOrEqualZero) ||
		errors.Is(err, workDomain.ErrInvalidWorkPriceValue) ||
		errors.Is(err, supplyDomain.ErrSupplyOutOfStock) ||
		errors.Is(err, reservationDomain.ErrInvalidReservationQuantity)
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
