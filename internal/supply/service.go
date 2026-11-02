package supply

import (
	"context"
	"math"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/interfaces"
	"golang.org/x/sync/errgroup"
)

type supplyService struct {
	uow         uow.Executor
	repo        interfaces.SupplyRepository
	idGenerator id.Generator
}

func NewService(uow uow.Executor, repo interfaces.SupplyRepository, idGenerator id.Generator) *supplyService {
	return &supplyService{uow, repo, idGenerator}
}

func (svc *supplyService) Create(ctx context.Context, supplyAdapter adapters.CreateSupply) (adapters.SupplyResponse, error) {
	supply := domain.NewSupply(
		svc.idGenerator.NewUUID(),
		supplyAdapter.Name,
		supplyAdapter.Description,
		supplyAdapter.UnitPrice,
		supplyAdapter.StockQuantity)

	if err := supply.Validate(); err != nil {
		return adapters.SupplyResponse{}, err
	}

	err := svc.uow.Execute(ctx, func(transactionContext context.Context) error {
		return svc.repo.Save(transactionContext, supply)
	})
	if err != nil {
		return adapters.SupplyResponse{}, err
	}

	return adapters.SupplyDomainToResponse(*supply), nil
}

// Update applies an optimistic-lock update: the write succeeds only when the
// stored version equals the version the caller read (supplyAdapter.Version).
// Stock is therefore never overwritten with a value read before a concurrent
// reservation (see ADR 0002).
func (svc *supplyService) Update(ctx context.Context, id string, supplyAdapter adapters.UpdateSupply) (adapters.SupplyResponse, error) {
	if id == "" {
		return adapters.SupplyResponse{}, domain.ErrInvalidSupplyID
	}

	if supplyAdapter.Version < 1 {
		return adapters.SupplyResponse{}, domain.ErrSupplyVersionRequired
	}

	supply := domain.NewSupply(
		id,
		supplyAdapter.Name,
		supplyAdapter.Description,
		supplyAdapter.UnitPrice,
		supplyAdapter.StockQuantity)

	if err := supply.Validate(); err != nil {
		return adapters.SupplyResponse{}, err
	}

	err := svc.uow.Execute(ctx, func(transactionContext context.Context) error {
		return svc.repo.UpdateWithVersion(transactionContext, supply, supplyAdapter.Version)
	})
	if err != nil {
		return adapters.SupplyResponse{}, err
	}

	return adapters.SupplyDomainToResponse(*supply), nil
}

func (svc *supplyService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrInvalidSupplyID
	}

	return svc.uow.Execute(ctx, func(transactionContext context.Context) error {
		return svc.repo.Delete(transactionContext, id)
	})
}

func (svc *supplyService) List(ctx context.Context, params adapters.ListSuppliesParams) (adapters.PaginatedSupplyResponse, error) {
	var totalItems int64
	var items []domain.Supply
	eg, groupCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		count, err := svc.repo.Count(groupCtx, params)
		if err != nil {
			return err
		}
		totalItems = count
		return nil
	})

	eg.Go(func() error {
		supplies, err := svc.repo.Search(groupCtx, params)
		if err != nil {
			return err
		}
		items = supplies
		return nil
	})

	if err := eg.Wait(); err != nil {
		return adapters.PaginatedSupplyResponse{}, err
	}

	return adapters.PaginatedSupplyResponse{
		Items:      adapters.SuppliesDomainToResponse(items),
		TotalItems: totalItems,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: int64(math.Max(float64(totalItems/params.PageSize), 1)),
	}, nil
}

func (svc *supplyService) FindById(ctx context.Context, id string) (domain.Supply, error) {
	if id == "" {
		return domain.Supply{}, domain.ErrInvalidSupplyID
	}

	supplies, err := svc.repo.Search(ctx, adapters.ListSuppliesParams{ID: id, Page: 1, PageSize: 1})
	if err != nil {
		return domain.Supply{}, err
	}

	if len(supplies) == 0 {
		return domain.Supply{}, domain.ErrSupplyNotFound
	}

	return supplies[0], nil
}

func (svc *supplyService) DecrementStockQuantity(ctx context.Context, id string, amount int) error {
	if id == "" {
		return domain.ErrInvalidSupplyID
	}

	if amount <= 0 {
		return domain.ErrInvalidSupplyAmount
	}

	list, err := svc.repo.Search(ctx, adapters.ListSuppliesParams{
		PageSize: 1,
		Page:     1,
		ID:       id,
	})
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return domain.ErrSupplyNotFound
	}

	supply := list[0]
	supply.StockQuantity -= amount

	err = svc.uow.Execute(ctx, func(transactionContext context.Context) error {
		return svc.repo.Save(transactionContext, &supply)
	})
	if err != nil {
		return err
	}

	return nil
}

func (svc *supplyService) IncrementStockQuantity(ctx context.Context, id string, amount int) error {
	if id == "" {
		return domain.ErrInvalidSupplyID
	}

	if amount <= 0 {
		return domain.ErrInvalidSupplyAmount
	}

	list, err := svc.repo.Search(ctx, adapters.ListSuppliesParams{
		PageSize: 1,
		Page:     1,
		ID:       id,
	})
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return domain.ErrSupplyNotFound
	}

	supply := list[0]
	supply.StockQuantity += amount

	err = svc.uow.Execute(ctx, func(transactionContext context.Context) error {
		return svc.repo.Save(transactionContext, &supply)
	})
	if err != nil {
		return err
	}

	return nil
}
