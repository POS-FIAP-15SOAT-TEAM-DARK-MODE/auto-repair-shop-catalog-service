package work

import (
	"context"
	"fmt"
	"math"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/interfaces"
	"golang.org/x/sync/errgroup"
)

type workService struct {
	uow  uow.Executor
	repo interfaces.WorkRepository
}

func NewService(uow uow.Executor, repo interfaces.WorkRepository) *workService {
	return &workService{uow: uow, repo: repo}
}

func (s *workService) Create(ctx context.Context, req adapters.CreateWork) (adapters.WorkResponse, error) {
	if err := domain.ValidWorkStatusStringValue(req.Status); err != nil {
		return adapters.WorkResponse{}, err
	}

	work, err := domain.NewWork(req.Name, req.Description, req.Price, domain.StringToWorkStatus(req.Status))
	if err != nil {
		return adapters.WorkResponse{}, fmt.Errorf("%w: %w", domain.ErrInvalidWorkPriceValue, err)
	}

	if err := work.Validate(); err != nil {
		return adapters.WorkResponse{}, err
	}

	if err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		return s.repo.Save(txCtx, work)
	}); err != nil {
		return adapters.WorkResponse{}, err
	}

	return adapters.WorkDomainToResponse(*work), nil
}

func (s *workService) Update(ctx context.Context, id string, req adapters.CreateWork) (adapters.WorkResponse, error) {
	if id == "" {
		return adapters.WorkResponse{}, domain.ErrInvalidWorkId
	}

	if err := domain.ValidWorkStatusStringValue(req.Status); err != nil {
		return adapters.WorkResponse{}, err
	}

	work, err := domain.NewWork(req.Name, req.Description, req.Price, domain.StringToWorkStatus(req.Status))
	if err != nil {
		return adapters.WorkResponse{}, fmt.Errorf("%w: %w", domain.ErrInvalidWorkPriceValue, err)
	}
	work.ID = id

	if err := work.Validate(); err != nil {
		return adapters.WorkResponse{}, err
	}

	if err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		return s.repo.Save(txCtx, work)
	}); err != nil {
		return adapters.WorkResponse{}, err
	}

	return adapters.WorkDomainToResponse(*work), nil
}

func (s *workService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrInvalidWorkId
	}

	if err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		return s.repo.Delete(txCtx, id)
	}); err != nil {
		return err
	}

	return nil
}

func (s *workService) List(ctx context.Context, params adapters.ListWorksParams) (adapters.PaginatedWorkResponse, error) {
	var totalItems int64
	var items []domain.Work
	eg, groupCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		count, err := s.repo.Count(groupCtx, params)
		if err != nil {
			return err
		}
		totalItems = count
		return nil
	})

	eg.Go(func() error {
		list, err := s.repo.Search(groupCtx, params)
		if err != nil {
			return err
		}
		items = list
		return nil
	})

	if err := eg.Wait(); err != nil {
		return adapters.PaginatedWorkResponse{}, fmt.Errorf("fail to list works: %w", err)
	}

	totalPages := int64(math.Max(1, math.Ceil(float64(totalItems)/float64(params.PageSize))))
	return adapters.PaginatedWorkResponse{
		Items:      adapters.WorksDomainToResponse(items),
		TotalItems: totalItems,
		TotalPages: totalPages,
		PageSize:   params.PageSize,
		Page:       params.Page,
	}, nil
}
