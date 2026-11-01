package supply_test

import (
	"context"
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	uowMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/interfaces/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// -------------------------
// Create
// -------------------------

func TestService_Create_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	s := adapters.CreateSupply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
	}

	exec.
		EXPECT().
		Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			if err := steps[0](ctx); err != nil {
				t.Fatalf("step returned error: %v", err)
			}
			return nil
		})

	repo.
		EXPECT().
		Save(mock.Anything, mock.AnythingOfType("*domain.Supply")).
		Return(nil)

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator("test"))

	res, err := appService.Create(ctx, s)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assert.Equal(t, s.Name, res.Name)
	assert.Equal(t, s.Description, res.Description)
	assert.Equal(t, s.UnitPrice, res.UnitPrice)
	assert.Equal(t, s.StockQuantity, res.StockQuantity)
	assert.Equal(t, "test", res.ID)
	assert.Equal(t, 1, res.Version)
}

func TestService_Create_ValidationError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	s := adapters.CreateSupply{} // invalid

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator("test"))

	_, err := appService.Create(ctx, s)
	assert.ErrorContains(t, err, domain.ErrInvalidSupplyName.Error())
	assert.ErrorContains(t, err, domain.ErrInvalidSupplyDescription.Error())
	assert.ErrorContains(t, err, domain.ErrInvalidSupplyUnitPrice.Error())

	exec.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestService_Create_UnitOfWorkError(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	s := adapters.CreateSupply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
	}

	expectedErr := errors.New("uow execute failed")

	exec.
		EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedErr)

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator("test"))

	_, err := appService.Create(ctx, s)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestService_Create_RepositoryErrorViaExecutor(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	s := adapters.CreateSupply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
	}

	expectedRepoErr := errors.New("db error")

	repo.
		EXPECT().
		Save(mock.Anything, mock.AnythingOfType("*domain.Supply")).
		Return(expectedRepoErr)

	exec.
		EXPECT().
		Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			return steps[0](ctx)
		})

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator("test"))

	_, err := appService.Create(ctx, s)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedRepoErr) {
		t.Fatalf("expected error to wrap repo error %v, got %v", expectedRepoErr, err)
	}
}

// -------------------------
// List
// -------------------------

func TestService_List_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	params := adapters.ListSuppliesParams{
		Page:     1,
		PageSize: 10,
	}

	expectedItems := []domain.Supply{
		{ID: "1", Name: "Brake Pad", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10},
	}
	var expectedTotal int64 = 1

	repo.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return(expectedTotal, nil)

	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return(expectedItems, nil)

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator("1"))

	got, err := appService.List(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.TotalItems != expectedTotal {
		t.Fatalf("expected total %d, got %d", expectedTotal, got.TotalItems)
	}

	if len(got.Items) != len(expectedItems) {
		t.Fatalf("expected %d items, got %d", len(expectedItems), len(got.Items))
	}

	if got.Page != params.Page {
		t.Fatalf("expected page %d, got %d", params.Page, got.Page)
	}

	if got.PageSize != params.PageSize {
		t.Fatalf("expected page size %d, got %d", params.PageSize, got.PageSize)
	}

	if got.TotalPages != 1 {
		t.Fatalf("expected 1 total page, got %d", got.TotalPages)
	}

	assert.ElementsMatch(t, adapters.SuppliesDomainToResponse(expectedItems), got.Items)
}

func TestService_List_UoWError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	params := adapters.ListSuppliesParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("uow execute failed")

	repo.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return(int64(0), expectedErr)

	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return(nil, nil)

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator("1"))

	_, err := appService.List(ctx, params)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestService_List_CountError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	params := adapters.ListSuppliesParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("count db error")

	repo.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return(int64(0), expectedErr).Maybe()

	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return([]domain.Supply{}, nil).Maybe()

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator("1"))

	_, err := appService.List(ctx, params)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestService_List_SearchError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	params := adapters.ListSuppliesParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("search db error")

	repo.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return(int64(0), nil).Maybe()

	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return(nil, expectedErr).Maybe()

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator("1"))

	_, err := appService.List(ctx, params)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestService_List_EmptyResult(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	params := adapters.ListSuppliesParams{Page: 1, PageSize: 10}

	repo.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return(int64(0), nil)

	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return([]domain.Supply{}, nil)

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator("1"))

	got, err := appService.List(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.TotalItems != 0 {
		t.Fatalf("expected total 0, got %d", got.TotalItems)
	}

	if len(got.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(got.Items))
	}

	if got.TotalPages != 1 {
		t.Fatalf("expected 1 total pages, got %d", got.TotalPages)
	}
}

// -------------------------
// Update
// -------------------------

func TestService_Update_Success(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)
	validId := "supply-id"

	s := adapters.CreateSupply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
	}

	exec.EXPECT().
		Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
			return steps[0](ctx)
		})

	repo.EXPECT().Search(mock.Anything, mock.Anything).Return([]domain.Supply{
		*domain.NewSupply(validId, s.Name, s.Description, s.UnitPrice, s.StockQuantity),
	}, nil)

	repo.EXPECT().
		Save(mock.Anything, mock.AnythingOfType("*domain.Supply")).
		Return(nil)

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator(validId))

	res, err := appService.Update(ctx, validId, s)
	assert.NoError(t, err)
	assert.Equal(t, s.Name, res.Name)
	assert.Equal(t, s.Description, res.Description)
	assert.Equal(t, s.UnitPrice, res.UnitPrice)
	assert.Equal(t, s.StockQuantity, res.StockQuantity)
	assert.Equal(t, validId, res.ID)
	assert.Equal(t, 2, res.Version)
}

func TestService_Update_ValidationError(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)
	validId := "supply-id"

	s := adapters.CreateSupply{} // invalid

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator(validId))

	_, err := appService.Update(ctx, validId, s)
	assert.ErrorContains(t, err, domain.ErrInvalidSupplyName.Error())
	assert.ErrorContains(t, err, domain.ErrInvalidSupplyDescription.Error())
	assert.ErrorContains(t, err, domain.ErrInvalidSupplyUnitPrice.Error())

	exec.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestService_Update_UnitOfWorkError(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	s := adapters.CreateSupply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
	}

	repo.EXPECT().Search(mock.Anything, mock.Anything).Return([]domain.Supply{
		*domain.NewSupply("test", s.Name, s.Description, s.UnitPrice, s.StockQuantity),
	}, nil)

	expectedErr := errors.New("uow execute failed")
	exec.
		EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedErr)

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator("test"))
	_, err := appService.Update(ctx, "test", s)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

// -------------------------
// Delete
// -------------------------

func TestService_Delete_Success(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	testId := "supply-id"

	exec.EXPECT().
		Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
			return steps[0](ctx)
		})

	repo.EXPECT().Delete(mock.Anything, testId).Return(nil)

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator(testId))
	err := appService.Delete(ctx, testId)
	assert.NoError(t, err)
}

func TestService_Delete_UnitOfWorkError(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	testId := "supply-id"
	expectedErr := errors.New("delete failed")

	exec.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedErr)

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator(testId))
	err := appService.Delete(ctx, testId)
	assert.Error(t, err, expectedErr)
}

func TestService_Delete_RepoError(t *testing.T) {
	ctx := context.Background()
	exec := uowMocks.NewExecutor(t)
	repo := mocks.NewSupplyRepository(t)

	testId := "supply-id"
	expectedErr := errors.New("delete failed")

	exec.EXPECT().
		Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
			return steps[0](ctx)
		})
	repo.EXPECT().Delete(mock.Anything, testId).Return(expectedErr)

	appService := supply.NewService(exec, repo, id.NewNoopIDGenerator(testId))
	err := appService.Delete(ctx, testId)
	assert.Error(t, err, expectedErr)
}

// -------------------------
// FindById
// -------------------------

func TestSupplyService_FindById(t *testing.T) {
	t.Parallel()

	validID := "abc-123"
	someSupply := *domain.NewSupply(
		validID,
		"Brake Pad",
		"High performance brake pad",
		decimal.NewFromFloat(49.99),
		10,
	)
	repoErr := errors.New("db connection failed")
	searchParams := adapters.ListSuppliesParams{ID: validID, Page: 1, PageSize: 1}

	tests := []struct {
		name       string
		id         string
		mockSetup  func(repo *mocks.SupplyRepository)
		wantSupply domain.Supply
		wantErr    error
	}{
		{
			name: "empty id returns ErrInvalidSupplyId without calling repo",
			id:   "",
			mockSetup: func(repo *mocks.SupplyRepository) {
				repo.AssertNumberOfCalls(t, "Search", 0)
			},
			wantSupply: domain.Supply{},
			wantErr:    domain.ErrInvalidSupplyId,
		},
		{
			name: "repo returns a supply, found successfully",
			id:   validID,
			mockSetup: func(repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{someSupply}, nil).
					Once()
			},
			wantSupply: someSupply,
			wantErr:    nil,
		},
		{
			name: "repo returns empty slice, supply not found",
			id:   "abc-123",
			mockSetup: func(repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{}, nil).
					Once()
			},
			wantSupply: domain.Supply{},
			wantErr:    domain.ErrSupplyNotFound,
		},
		{
			name: "repo returns an error, propagated as-is",
			id:   "abc-123",
			mockSetup: func(repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{}, repoErr).
					Once()
			},
			wantSupply: domain.Supply{},
			wantErr:    repoErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			exec := uowMocks.NewExecutor(t)
			repo := mocks.NewSupplyRepository(t)
			tc.mockSetup(repo)

			svc := supply.NewService(exec, repo, nil)
			got, err := svc.FindById(context.Background(), tc.id)

			assert.Equal(t, err, tc.wantErr)
			assert.Equal(t, tc.wantSupply, got)
		})
	}
}

// -------------------------
// DecrementStockQuantity
// -------------------------

func TestSupplyService_DecrementStockQuantity(t *testing.T) {
	t.Parallel()

	validID := "abc-123"
	someSupply := *domain.NewSupply(
		validID,
		"Brake Pad",
		"High performance brake pad",
		decimal.NewFromFloat(49.99),
		10,
	)
	repoErr := errors.New("db connection failed")
	searchParams := adapters.ListSuppliesParams{ID: validID, Page: 1, PageSize: 1}

	tests := []struct {
		name      string
		id        string
		amount    int
		mockSetup func(uow *uowMocks.Executor, repo *mocks.SupplyRepository)
		wantErr   error
	}{
		{
			name: "empty id returns ErrInvalidSupplyId without calling repo",
			id:   "",
			mockSetup: func(uow *uowMocks.Executor, repo *mocks.SupplyRepository) {
				uow.AssertNumberOfCalls(t, "Execute", 0)
				repo.AssertNumberOfCalls(t, "Search", 0)
				repo.AssertNumberOfCalls(t, "Save", 0)
			},
			wantErr: domain.ErrInvalidSupplyId,
		},
		{
			name: "invalid amount returns ErrInvalidSupplyAmount without calling repo",
			id:   validID,
			mockSetup: func(uow *uowMocks.Executor, repo *mocks.SupplyRepository) {
				uow.AssertNumberOfCalls(t, "Execute", 0)
				repo.AssertNumberOfCalls(t, "Search", 0)
				repo.AssertNumberOfCalls(t, "Save", 0)
			},
			wantErr: domain.ErrInvalidSupplyAmount,
		},
		{
			name:   "repo returns empty slice, supply not found",
			id:     validID,
			amount: 1,
			mockSetup: func(uow *uowMocks.Executor, repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{}, nil).
					Once()
				uow.AssertNumberOfCalls(t, "Execute", 0)
				repo.AssertNumberOfCalls(t, "Save", 0)
			},
			wantErr: domain.ErrSupplyNotFound,
		},
		{
			name:   "search repo returns an error, propagated as-is",
			id:     validID,
			amount: 1,
			mockSetup: func(uow *uowMocks.Executor, repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{}, repoErr).
					Once()
				uow.AssertNumberOfCalls(t, "Execute", 0)
				repo.AssertNumberOfCalls(t, "Save", 0)
			},
			wantErr: repoErr,
		},
		{
			name:   "unit of work returns an error, propagated as-is",
			id:     validID,
			amount: 1,
			mockSetup: func(uow *uowMocks.Executor, repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{someSupply}, nil).
					Once()
				uow.EXPECT().
					Execute(mock.Anything, mock.Anything).
					Return(repoErr)
				repo.AssertNumberOfCalls(t, "Save", 0)
			},
			wantErr: repoErr,
		},
		{
			name:   "save repo returns an error, propagated as-is",
			id:     validID,
			amount: 1,
			mockSetup: func(exec *uowMocks.Executor, repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{someSupply}, nil).
					Once()
				exec.EXPECT().
					Execute(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
						return steps[0](ctx)
					})
				s := someSupply
				s.StockQuantity -= 1
				repo.EXPECT().
					Save(mock.Anything, &s).
					Return(repoErr).
					Once()
			},
			wantErr: repoErr,
		},
		{
			name:   "decrement stock quantity",
			id:     validID,
			amount: 1,
			mockSetup: func(exec *uowMocks.Executor, repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{someSupply}, nil).
					Once()
				exec.EXPECT().
					Execute(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
						return steps[0](ctx)
					})
				s := someSupply
				s.StockQuantity -= 1
				repo.EXPECT().
					Save(mock.Anything, &s).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			exec := uowMocks.NewExecutor(t)
			repo := mocks.NewSupplyRepository(t)
			tc.mockSetup(exec, repo)

			svc := supply.NewService(exec, repo, nil)
			err := svc.DecrementStockQuantity(context.Background(), tc.id, tc.amount)

			assert.Equal(t, err, tc.wantErr)
		})
	}
}

// -------------------------
// IncrementStockQuantity
// -------------------------

func TestSupplyService_IncrementStockQuantity(t *testing.T) {
	t.Parallel()

	validID := "abc-123"
	someSupply := *domain.NewSupply(
		validID,
		"Brake Pad",
		"High performance brake pad",
		decimal.NewFromFloat(49.99),
		10,
	)
	repoErr := errors.New("db connection failed")
	searchParams := adapters.ListSuppliesParams{ID: validID, Page: 1, PageSize: 1}

	tests := []struct {
		name      string
		id        string
		amount    int
		mockSetup func(uow *uowMocks.Executor, repo *mocks.SupplyRepository)
		wantErr   error
	}{
		{
			name: "empty id returns ErrInvalidSupplyId without calling repo",
			id:   "",
			mockSetup: func(uow *uowMocks.Executor, repo *mocks.SupplyRepository) {
				uow.AssertNumberOfCalls(t, "Execute", 0)
				repo.AssertNumberOfCalls(t, "Search", 0)
				repo.AssertNumberOfCalls(t, "Save", 0)
			},
			wantErr: domain.ErrInvalidSupplyId,
		},
		{
			name: "invalid amount returns ErrInvalidSupplyAmount without calling repo",
			id:   validID,
			mockSetup: func(uow *uowMocks.Executor, repo *mocks.SupplyRepository) {
				uow.AssertNumberOfCalls(t, "Execute", 0)
				repo.AssertNumberOfCalls(t, "Search", 0)
				repo.AssertNumberOfCalls(t, "Save", 0)
			},
			wantErr: domain.ErrInvalidSupplyAmount,
		},
		{
			name:   "repo returns empty slice, supply not found",
			id:     validID,
			amount: 1,
			mockSetup: func(uow *uowMocks.Executor, repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{}, nil).
					Once()
				uow.AssertNumberOfCalls(t, "Execute", 0)
				repo.AssertNumberOfCalls(t, "Save", 0)
			},
			wantErr: domain.ErrSupplyNotFound,
		},
		{
			name:   "search repo returns an error, propagated as-is",
			id:     validID,
			amount: 1,
			mockSetup: func(uow *uowMocks.Executor, repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{}, repoErr).
					Once()
				uow.AssertNumberOfCalls(t, "Execute", 0)
				repo.AssertNumberOfCalls(t, "Save", 0)
			},
			wantErr: repoErr,
		},
		{
			name:   "unit of work returns an error, propagated as-is",
			id:     validID,
			amount: 1,
			mockSetup: func(uow *uowMocks.Executor, repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{someSupply}, nil).
					Once()
				uow.EXPECT().
					Execute(mock.Anything, mock.Anything).
					Return(repoErr)
				repo.AssertNumberOfCalls(t, "Save", 0)
			},
			wantErr: repoErr,
		},
		{
			name:   "save repo returns an error, propagated as-is",
			id:     validID,
			amount: 1,
			mockSetup: func(exec *uowMocks.Executor, repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{someSupply}, nil).
					Once()
				exec.EXPECT().
					Execute(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
						return steps[0](ctx)
					})
				s := someSupply
				s.StockQuantity += 1
				repo.EXPECT().
					Save(mock.Anything, &s).
					Return(repoErr).
					Once()
			},
			wantErr: repoErr,
		},
		{
			name:   "decrement stock quantity",
			id:     validID,
			amount: 1,
			mockSetup: func(exec *uowMocks.Executor, repo *mocks.SupplyRepository) {
				repo.EXPECT().
					Search(mock.Anything, searchParams).
					Return([]domain.Supply{someSupply}, nil).
					Once()
				exec.EXPECT().
					Execute(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
						return steps[0](ctx)
					})
				s := someSupply
				s.StockQuantity += 1
				repo.EXPECT().
					Save(mock.Anything, &s).
					Return(nil).
					Once()
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			exec := uowMocks.NewExecutor(t)
			repo := mocks.NewSupplyRepository(t)
			tc.mockSetup(exec, repo)

			svc := supply.NewService(exec, repo, nil)
			err := svc.IncrementStockQuantity(context.Background(), tc.id, tc.amount)

			assert.Equal(t, err, tc.wantErr)
		})
	}
}
