package routing

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRespondError_BodyNeverCarriesInfraCause(t *testing.T) {
	gin.SetMode(gin.TestMode)
	driverErr := errors.New(`pq: duplicate key value violates unique constraint "supplies_pkey"`)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{"unknown 500", fmt.Errorf("repo: %w", driverErr), http.StatusInternalServerError, `{"code":500,"error":"internal server error"}`},
		{"infra 409", fmt.Errorf("%w: %v", app.ErrDataConflict, driverErr), http.StatusConflict, `{"code":409,"error":"conflict"}`},
		{"infra 422", fmt.Errorf("%w: %v", app.ErrDataViolation, driverErr), http.StatusUnprocessableEntity, `{"code":422,"error":"invalid data"}`},
		{"domain 409 keeps message", supplyDomain.ErrSupplyVersionConflict, http.StatusConflict, `{"code":409,"error":"` + supplyDomain.ErrSupplyVersionConflict.Error() + `"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			respondError(c, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.JSONEq(t, tt.wantBody, rec.Body.String())
			assert.NotContains(t, rec.Body.String(), "pq:")
		})
	}
}
