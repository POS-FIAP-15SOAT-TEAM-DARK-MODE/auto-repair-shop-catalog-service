package supply_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/routing"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/repository"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The PUT route goes through the real controller, service and in-memory repository.
func newUpdateRouter(t *testing.T, stored *domain.Supply) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	repo := repository.NewInMemory()
	require.NoError(t, repo.Save(context.Background(), stored))
	svc := supply.NewService(uow.NewInMemoryUoW(), repo, id.NewNoopIDGenerator("unused"))
	router := gin.New()
	router.PUT("/v1/supplies/:id", routing.GinHandler(supply.NewController(svc).Update))
	return router
}

func putSupply(router http.Handler, supplyID, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/v1/supplies/"+supplyID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestPutSupply_OptimisticLock(t *testing.T) {
	stored := domain.NewSupply("supply-1", "Brake Pad", "High performance brake pad", decimal.NewFromFloat(49.99), 10)
	router := newUpdateRouter(t, stored) // stored version is 1

	withVersion := func(v string) string {
		return `{"name":"Brake Pad","description":"High performance brake pad","unitPrice":"49.99","stockQuantity":10` + v + `}`
	}

	t.Run("missing version is a bad request", func(t *testing.T) {
		rec := putSupply(router, "supply-1", withVersion(""))
		assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), domain.ErrSupplyVersionRequired.Error())
	})

	t.Run("version below one is a bad request", func(t *testing.T) {
		rec := putSupply(router, "supply-1", withVersion(`,"version":0`))
		assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	})

	t.Run("stale version conflicts with a generic message", func(t *testing.T) {
		rec := putSupply(router, "supply-1", withVersion(`,"version":7`))
		assert.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), domain.ErrSupplyVersionConflict.Error())
	})

	t.Run("current version updates and returns the new version", func(t *testing.T) {
		rec := putSupply(router, "supply-1", withVersion(`,"version":1`))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var body struct {
			Version int `json:"version"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, 2, body.Version)
	})

	t.Run("unknown supply is not found", func(t *testing.T) {
		rec := putSupply(router, "missing", withVersion(`,"version":1`))
		assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
	})
}
