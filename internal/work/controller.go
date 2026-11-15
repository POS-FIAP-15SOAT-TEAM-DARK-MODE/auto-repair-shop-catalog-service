package work

import (
	"context"
	"net/http"
	"strconv"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/interfaces"
	"go.uber.org/zap"
)

const (
	idParam       = "id"
	pageParam     = "page"
	pageSizeParam = "pageSize"
	statusParam   = "status"
	defaultPage   = 1
	defaultSize   = 10
)

type workController struct {
	svc interfaces.WorkService
}

func NewController(svc interfaces.WorkService) *workController {
	return &workController{svc: svc}
}

func (ctrl *workController) Create(ctx context.Context, req *http.Request) (adapters.WorkResponse, error) {
	var dto adapters.CreateWork
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Failed to bind work creation payload",
			zap.String("operation", "create_work"), zap.Error(err))
		return adapters.WorkResponse{}, err
	}

	resp, err := ctrl.svc.Create(ctx, dto)
	if err != nil {
		logger.Of(ctx).Debug("Work creation failed in service layer",
			zap.String("operation", "create_work"), zap.Error(err))
		return adapters.WorkResponse{}, err
	}

	return resp, nil
}

func (ctrl *workController) List(ctx context.Context, req *http.Request) (adapters.PaginatedWorkResponse, error) {
	q := req.URL.Query()
	page := int64(defaultPage)
	pageSize := int64(defaultSize)

	if p := q.Get(pageParam); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = int64(v)
		}
	}
	if ps := q.Get(pageSizeParam); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = int64(v)
		}
	}

	status := q.Get(statusParam)
	if status != "" {
		if err := domain.ValidWorkStatusStringValue(status); err != nil {
			status = ""
		}
	}

	resp, err := ctrl.svc.List(ctx, adapters.ListWorksParams{Page: page, PageSize: pageSize, Status: status})
	if err != nil {
		logger.Of(ctx).Debug("Work list failed in service layer",
			zap.String("operation", "list_works"), zap.Error(err))
		return adapters.PaginatedWorkResponse{}, err
	}

	return resp, nil
}

func (ctrl *workController) Update(ctx context.Context, req *http.Request) (adapters.WorkResponse, error) {
	id, _ := ctx.Value(idParam).(string)
	if id == "" {
		return adapters.WorkResponse{}, domain.ErrInvalidWorkId
	}

	var dto adapters.CreateWork
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Failed to bind work update payload",
			zap.String("operation", "update_work"), zap.Error(err))
		return adapters.WorkResponse{}, err
	}

	resp, err := ctrl.svc.Update(ctx, id, dto)
	if err != nil {
		logger.Of(ctx).Debug("Work update failed in service layer",
			zap.String("operation", "update_work"), zap.Error(err))
		return adapters.WorkResponse{}, err
	}

	return resp, nil
}

func (ctrl *workController) Delete(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(idParam).(string)
	if id == "" {
		return domain.ErrInvalidWorkId
	}

	if err := ctrl.svc.Delete(ctx, id); err != nil {
		logger.Of(ctx).Debug("Work deletion failed in service layer",
			zap.String("operation", "delete_work"), zap.Error(err))
		return err
	}

	return nil
}
