package work_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	uowPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/interfaces/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func workInMemoryRepo() interfaces.WorkRepository {
	return repository.NewInMemory()
}

func workInMemoryUoW() uowPkg.Executor {
	return uowPkg.NewInMemoryUoW()
}

// ─── Create ───────────────────────────────────────────────────────────────────

func TestController_Create_Success(t *testing.T) {
	svcMock := mocks.NewWorkService(t)
	ctrl := work.NewController(svcMock)

	body, _ := json.Marshal(map[string]string{
		"name": "Oil Change", "description": "Synthetic oil change service",
		"price": "199.90", "status": "ACTIVE",
	})
	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	expected := adapters.WorkResponse{ID: "uuid-1", Name: "Oil Change", Price: "199.90", Status: "ACTIVE"}
	svcMock.EXPECT().Create(mock.Anything, mock.AnythingOfType("adapters.CreateWork")).Return(expected, nil)

	resp, err := ctrl.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, resp.ID)
	assert.Equal(t, expected.Name, resp.Name)
}

func TestController_Create_InvalidStatus(t *testing.T) {
	repo := workInMemoryRepo()
	uow := workInMemoryUoW()
	svc := work.NewService(uow, repo)
	ctrl := work.NewController(svc)

	body, _ := json.Marshal(map[string]string{
		"name": "Oil Change", "description": "Valid description here",
		"price": "10.00", "status": "INVALID",
	})
	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	_, err := ctrl.Create(context.Background(), req)

	assert.ErrorIs(t, err, domain.ErrInvalidWorkStatusValue)
}

func TestController_Create_InvalidJSON(t *testing.T) {
	svcMock := mocks.NewWorkService(t)
	ctrl := work.NewController(svcMock)

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	_, err := ctrl.Create(context.Background(), req)

	assert.Error(t, err)
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestController_List_Success(t *testing.T) {
	svcMock := mocks.NewWorkService(t)
	ctrl := work.NewController(svcMock)

	req := httptest.NewRequest(http.MethodGet, "/works?page=1&pageSize=5&status=ACTIVE", nil)

	expected := adapters.PaginatedWorkResponse{TotalItems: 1, Page: 1, PageSize: 5}
	svcMock.EXPECT().
		List(mock.Anything, adapters.ListWorksParams{Page: 1, PageSize: 5, Status: "ACTIVE"}).
		Return(expected, nil)

	resp, err := ctrl.List(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expected.TotalItems, resp.TotalItems)
}

func TestController_List_DefaultParams(t *testing.T) {
	svcMock := mocks.NewWorkService(t)
	ctrl := work.NewController(svcMock)

	req := httptest.NewRequest(http.MethodGet, "/works", nil)

	svcMock.EXPECT().
		List(mock.Anything, adapters.ListWorksParams{Page: 1, PageSize: 10, Status: ""}).
		Return(adapters.PaginatedWorkResponse{}, nil)

	_, err := ctrl.List(context.Background(), req)

	assert.NoError(t, err)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TestController_Update_Success(t *testing.T) {
	svcMock := mocks.NewWorkService(t)
	ctrl := work.NewController(svcMock)

	body, _ := json.Marshal(map[string]string{
		"name": "Brake Service", "description": "Complete brake inspection and pad replacement",
		"price": "350.00", "status": "ACTIVE",
	})
	req := httptest.NewRequest(http.MethodPut, "/works/uuid-1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(context.Background(), "id", "uuid-1") //nolint:staticcheck

	expected := adapters.WorkResponse{ID: "uuid-1", Name: "Brake Service"}
	svcMock.EXPECT().Update(mock.Anything, "uuid-1", mock.AnythingOfType("adapters.CreateWork")).Return(expected, nil)

	resp, err := ctrl.Update(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "uuid-1", resp.ID)
}

func TestController_Update_MissingID(t *testing.T) {
	svcMock := mocks.NewWorkService(t)
	ctrl := work.NewController(svcMock)

	req := httptest.NewRequest(http.MethodPut, "/works/", nil)

	_, err := ctrl.Update(context.Background(), req)

	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestController_Delete_Success(t *testing.T) {
	svcMock := mocks.NewWorkService(t)
	ctrl := work.NewController(svcMock)

	req := httptest.NewRequest(http.MethodDelete, "/works/uuid-1", nil)
	ctx := context.WithValue(context.Background(), "id", "uuid-1") //nolint:staticcheck

	svcMock.EXPECT().Delete(mock.Anything, "uuid-1").Return(nil)

	err := ctrl.Delete(ctx, req)

	assert.NoError(t, err)
}

func TestController_Delete_MissingID(t *testing.T) {
	svcMock := mocks.NewWorkService(t)
	ctrl := work.NewController(svcMock)

	req := httptest.NewRequest(http.MethodDelete, "/works/", nil)

	err := ctrl.Delete(context.Background(), req)

	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

// ─── Integration (in-memory) ──────────────────────────────────────────────────

func TestController_Create_Integration(t *testing.T) {
	repo := workInMemoryRepo()
	uow := workInMemoryUoW()
	svc := work.NewService(uow, repo)
	ctrl := work.NewController(svc)

	body, _ := json.Marshal(map[string]string{
		"name": "Oil Change", "description": "Complete synthetic oil change",
		"price": "199.90", "status": "ACTIVE",
	})
	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := ctrl.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "Oil Change", resp.Name)
	assert.Equal(t, "ACTIVE", resp.Status)
}
