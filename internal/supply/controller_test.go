package supply_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/routing"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/interfaces/mocks"
)

// -------------------------
// Setup helpers
// -------------------------

func validSupplyDomain() *domain.Supply {
	return domain.NewSupply(
		uuid.New().String(),
		"Brake Pad",
		"High performance brake pad",
		decimal.NewFromFloat(49.99),
		10)
}

func validSupplyPayload() map[string]any {
	return map[string]any{
		"name":          "Brake Pad",
		"description":   "High performance brake pad",
		"unitPrice":     "49.99",
		"stockQuantity": 10,
	}
}

func toJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	return bytes.NewBuffer(b)
}

// -------------------------
// Create
// -------------------------

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		payload        any
		mockSetup      func(*mocks.SupplyService)
		expectedStatus int
	}{
		{
			name:    "success_creates_supply",
			payload: validSupplyPayload(),
			mockSetup: func(svc *mocks.SupplyService) {
				svc.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("adapters.CreateSupply")).
					Return(adapters.SupplyResponse{}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "service_error_on_first_call",
			payload: validSupplyPayload(),
			mockSetup: func(svc *mocks.SupplyService) {
				svc.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("adapters.CreateSupply")).
					Return(adapters.SupplyResponse{}, errors.New("internal error")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid_payload_returns_bad_request",
			payload:        `{"invalid-json":`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewSupplyService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			h := supply.NewController(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(http.MethodPost, "/", toJSON(t, tt.payload))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			routing.GinHandler(h.Create, http.StatusCreated)(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreate_ResponseBody(t *testing.T) {
	svc := mocks.NewSupplyService(t)
	svc.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("adapters.CreateSupply")).
		RunAndReturn(func(ctx context.Context, createSupply adapters.CreateSupply) (adapters.SupplyResponse, error) {
			return adapters.SupplyDomainToResponse(*validSupplyDomain()), nil
		})

	h := supply.NewController(svc)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	req := httptest.NewRequest(http.MethodPost, "/", toJSON(t, validSupplyPayload()))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	routing.GinHandler(h.Create, http.StatusCreated)(c)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "Brake Pad", body["name"])
	assert.Equal(t, "High performance brake pad", body["description"])
}

// -------------------------
// List
// -------------------------

func TestList(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockSetup      func(*mocks.SupplyService)
		expectedStatus int
	}{
		{
			name:  "success_returns_supplies",
			query: "?page=1&pageSize=10",
			mockSetup: func(svc *mocks.SupplyService) {
				svc.EXPECT().
					List(mock.Anything, mock.AnythingOfType("adapters.ListSuppliesParams")).
					Return(adapters.PaginatedSupplyResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "success_returns_empty_list",
			query: "?page=1&pageSize=10",
			mockSetup: func(svc *mocks.SupplyService) {
				svc.EXPECT().
					List(mock.Anything, mock.AnythingOfType("adapters.ListSuppliesParams")).
					Return(adapters.PaginatedSupplyResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "success_uses_default_params_when_no_query",
			query: "",
			mockSetup: func(svc *mocks.SupplyService) {
				svc.EXPECT().
					List(mock.Anything, mock.AnythingOfType("adapters.ListSuppliesParams")).
					Return(adapters.PaginatedSupplyResponse{}, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewSupplyService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			h := supply.NewController(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(http.MethodGet, "/supplies"+tt.query, nil)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			routing.GinHandler(h.List)(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestList_ResponseBody(t *testing.T) {
	supply1 := adapters.SupplyDomainToResponse(*validSupplyDomain())
	supply2 := adapters.SupplyDomainToResponse(*validSupplyDomain())

	svc := mocks.NewSupplyService(t)
	svc.EXPECT().
		List(mock.Anything, mock.AnythingOfType("adapters.ListSuppliesParams")).
		Return(adapters.PaginatedSupplyResponse{
			Items: []adapters.SupplyResponse{supply1, supply2},
		}, nil)

	h := supply.NewController(svc)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/supplies?page=1&pageSize=2", nil)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	routing.GinHandler(h.List)(c)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	items, ok := body["items"].([]any)
	assert.True(t, ok)
	assert.Len(t, items, 2)

	for _, item := range items {
		entry := item.(map[string]any)
		assert.NotEmpty(t, entry["id"])
		assert.Equal(t, "Brake Pad", entry["name"])
		assert.Equal(t, "High performance brake pad", entry["description"])
		assert.Equal(t, decimal.NewFromFloat(49.99).String(), entry["unit_price"].(string))
	}
}

func TestList_ResponseBody_Empty(t *testing.T) {
	svc := mocks.NewSupplyService(t)
	svc.EXPECT().
		List(mock.Anything, mock.AnythingOfType("adapters.ListSuppliesParams")).
		Return(adapters.PaginatedSupplyResponse{
			Items: []adapters.SupplyResponse{},
		}, nil)

	h := supply.NewController(svc)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/supplies?page=1&pageSize=10", nil)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	routing.GinHandler(h.List)(c)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	items, ok := body["items"].([]any)
	assert.True(t, ok)
	assert.Empty(t, items)
}

// -------------------------
// Update
// -------------------------

func TestUpdate(t *testing.T) {
	validID := uuid.New().String()

	tests := []struct {
		name           string
		id             string
		payload        any
		mockSetup      func(*mocks.SupplyService)
		expectedStatus int
	}{
		{
			name:    "success_updates_supply",
			id:      validID,
			payload: validSupplyPayload(),
			mockSetup: func(svc *mocks.SupplyService) {
				svc.EXPECT().
					Update(mock.Anything, validID, mock.AnythingOfType("adapters.CreateSupply")).
					Return(adapters.SupplyResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "service_error_returns_internal_error",
			id:      validID,
			payload: validSupplyPayload(),
			mockSetup: func(svc *mocks.SupplyService) {
				svc.EXPECT().
					Update(mock.Anything, validID, mock.AnythingOfType("adapters.CreateSupply")).
					Return(adapters.SupplyResponse{}, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid_payload_returns_bad_request",
			id:             validID,
			payload:        "invalid-json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewSupplyService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(svc)
			}

			h := supply.NewController(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Params = gin.Params{gin.Param{Key: "id", Value: validID}}
			url := fmt.Sprintf("/%s", tt.id)
			req := httptest.NewRequest(http.MethodPut, url, toJSON(t, tt.payload))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			routing.GinHandler(h.Update)(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUpdate_ResponseBody(t *testing.T) {
	validID := uuid.New().String()

	svc := mocks.NewSupplyService(t)
	svc.EXPECT().
		Update(mock.Anything, validID, mock.AnythingOfType("adapters.CreateSupply")).
		RunAndReturn(func(ctx context.Context, id string, createSupply adapters.CreateSupply) (adapters.SupplyResponse, error) {
			res := adapters.SupplyDomainToResponse(*validSupplyDomain())
			res.ID = id
			return res, nil
		})

	h := supply.NewController(svc)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{gin.Param{Key: "id", Value: validID}}

	url := fmt.Sprintf("/%s", validID)
	req := httptest.NewRequest(http.MethodPut, url, toJSON(t, validSupplyPayload()))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	routing.GinHandler(h.Update)(c)

	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, validID, body["id"])
	assert.Equal(t, "Brake Pad", body["name"])
}

// -------------------------
// Delete
// -------------------------

func TestDeleteWorkHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validID := "supply-id"
	svc := mocks.NewSupplyService(t)
	h := supply.NewController(svc)

	svc.EXPECT().Delete(mock.Anything, validID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{gin.Param{Key: "id", Value: validID}}
	url := fmt.Sprintf("/%s", validID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	routing.GinOnlyErrorHandler(h.Delete)(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteWorkHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validID := "supply-id"
	svc := mocks.NewSupplyService(t)
	h := supply.NewController(svc)

	svc.EXPECT().Delete(mock.Anything, validID).Return(errors.New("internal error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{gin.Param{Key: "id", Value: validID}}
	url := fmt.Sprintf("/%s", validID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	routing.GinOnlyErrorHandler(h.Delete)(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
