package work

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/interfaces"
)

// internalWorkService serves the read-only lookup used by other services over /internal.
type internalWorkService struct {
	uow  uow.Executor
	repo interfaces.WorkRepository
}

func NewInternalService(unit uow.Executor, repo interfaces.WorkRepository) *internalWorkService {
	return &internalWorkService{uow: unit, repo: repo}
}

func (s *internalWorkService) FindByID(ctx context.Context, id string) (adapters.InternalWorkResponse, error) {
	if id == "" {
		return adapters.InternalWorkResponse{}, domain.ErrInvalidWorkId
	}

	var w domain.Work
	if err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		var err error
		w, err = s.repo.FindByID(txCtx, id)
		return err
	}); err != nil {
		return adapters.InternalWorkResponse{}, err
	}
	return adapters.InternalWorkDomainToResponse(w), nil
}

type internalWorkController struct {
	svc interfaces.WorkInternalService
}

func NewInternalController(svc interfaces.WorkInternalService) *internalWorkController {
	return &internalWorkController{svc: svc}
}

func (c *internalWorkController) Get(ctx context.Context, _ *http.Request) (adapters.InternalWorkResponse, error) {
	id, _ := ctx.Value(idParam).(string)
	return c.svc.FindByID(ctx, id)
}
