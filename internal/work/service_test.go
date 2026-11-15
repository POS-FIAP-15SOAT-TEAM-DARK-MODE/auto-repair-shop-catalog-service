package work_test

import (
	"context"
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	uowMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func runUoW(ctx context.Context, steps ...uow.Step) error {
	for _, s := range steps {
		if err := s(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func TestService_Create_Success(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewWorkRepository(t)

	req := adapters.CreateWork{
		Name:        "Oil Change",
		Description: "Complete synthetic oil change",
		Price:       "199.90",
		Status:      "ACTIVE",
	}

	exec.EXPECT().Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(c context.Context, steps ...uow.Step) error { return runUoW(c, steps...) })
	repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.Work")).Return(nil)

	svc := work.NewService(exec, repo)
	res, err := svc.Create(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, req.Name, res.Name)
	assert.Equal(t, req.Description, res.Description)
	assert.Equal(t, "ACTIVE", res.Status)
	assert.NotEmpty(t, res.ID)
}

func TestService_Create_ValidationError_EmptyName(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewWorkRepository(t)

	req := adapters.CreateWork{
		Name:        "",
		Description: "Complete synthetic oil change",
		Price:       "199.90",
		Status:      "ACTIVE",
	}

	svc := work.NewService(exec, repo)
	_, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrEmptyWorkName)
	exec.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
}

func TestService_Create_ValidationError_ShortDescription(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewWorkRepository(t)

	req := adapters.CreateWork{Name: "Oil Change", Description: "Short", Price: "10.00", Status: "ACTIVE"}

	svc := work.NewService(exec, repo)
	_, err := svc.Create(ctx, req)

	assert.ErrorIs(t, err, domain.ErrWorkDescriptionShorterThenRequired)
}

func TestService_Create_UoWError(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewWorkRepository(t)
	uowErr := errors.New("db error")

	req := adapters.CreateWork{
		Name: "Oil Change", Description: "Complete synthetic oil change",
		Price: "199.90", Status: "ACTIVE",
	}

	exec.EXPECT().Execute(mock.Anything, mock.Anything).Return(uowErr)

	svc := work.NewService(exec, repo)
	_, err := svc.Create(ctx, req)

	assert.ErrorIs(t, err, uowErr)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TestService_Update_Success(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewWorkRepository(t)

	req := adapters.CreateWork{
		Name:        "Oil Change Pro",
		Description: "Complete synthetic oil change with filter",
		Price:       "249.90",
		Status:      "ACTIVE",
	}

	exec.EXPECT().Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(c context.Context, steps ...uow.Step) error { return runUoW(c, steps...) })
	repo.EXPECT().Save(mock.Anything, mock.AnythingOfType("*domain.Work")).Return(nil)

	svc := work.NewService(exec, repo)
	res, err := svc.Update(ctx, "some-uuid", req)

	assert.NoError(t, err)
	assert.Equal(t, "some-uuid", res.ID)
	assert.Equal(t, req.Name, res.Name)
}

func TestService_Update_EmptyID(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewWorkRepository(t)

	svc := work.NewService(exec, repo)
	_, err := svc.Update(ctx, "", adapters.CreateWork{})

	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func TestService_Delete_Success(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewWorkRepository(t)

	exec.EXPECT().Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(c context.Context, steps ...uow.Step) error { return runUoW(c, steps...) })
	repo.EXPECT().Delete(mock.Anything, "some-id").Return(nil)

	svc := work.NewService(exec, repo)
	err := svc.Delete(ctx, "some-id")

	assert.NoError(t, err)
}

func TestService_Delete_EmptyID(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewWorkRepository(t)

	svc := work.NewService(exec, repo)
	err := svc.Delete(ctx, "")

	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestService_List_Success(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewWorkRepository(t)

	params := adapters.ListWorksParams{Page: 1, PageSize: 10, Status: "ACTIVE"}

	repo.EXPECT().Count(mock.Anything, params).Return(int64(2), nil)
	repo.EXPECT().Search(mock.Anything, params).Return([]domain.Work{
		{ID: "1", Name: "Oil Change", Description: "Description here", Status: domain.ACTIVE},
		{ID: "2", Name: "Tire Rotation", Description: "Another description", Status: domain.ACTIVE},
	}, nil)

	svc := work.NewService(exec, repo)
	res, err := svc.List(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), res.TotalItems)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, int64(1), res.Page)
	assert.Equal(t, int64(10), res.PageSize)
}

func TestService_List_RepoError(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewWorkRepository(t)

	dbErr := errors.New("db error")
	params := adapters.ListWorksParams{Page: 1, PageSize: 10}

	repo.EXPECT().Count(mock.Anything, params).Return(int64(0), dbErr).Maybe()
	repo.EXPECT().Search(mock.Anything, params).Return(nil, dbErr).Maybe()

	svc := work.NewService(exec, repo)
	_, err := svc.List(ctx, params)

	assert.Error(t, err)
}
