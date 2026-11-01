package supply

import (
	"context"
	"net/http"
	"strconv"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/interfaces"
	"go.uber.org/zap"
)

const (
	idParam         = "id"
	versionParam    = "version"
	pageParam       = "page"
	pageSizeParam   = "pageSize"
	defaultPageSize = 10
	defaultPage     = 1
)

type supplyController struct {
	svc interfaces.SupplyService
}

func NewController(svc interfaces.SupplyService) *supplyController {
	return &supplyController{svc: svc}
}

func (ctrl *supplyController) Create(ctx context.Context, req *http.Request) (adapters.SupplyResponse, error) {
	dto := adapters.CreateSupply{}
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Failed to bind supply creation payload",
			zap.String("operation", "create_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		return adapters.SupplyResponse{}, err
	}

	resp, err := ctrl.svc.Create(ctx, dto)
	if err != nil {
		logger.Of(ctx).Debug("Supply creation failed in service layer",
			zap.String("operation", "create_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		return adapters.SupplyResponse{}, err
	}

	return resp, nil
}

func (ctrl *supplyController) List(ctx context.Context, req *http.Request) (adapters.PaginatedSupplyResponse, error) {
	params := req.URL.Query()
	pg := int64(defaultPage)
	pgSize := int64(defaultPageSize)
	if page := params.Get(pageParam); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > defaultPage {
			pg = int64(p)
		}
	}
	if pageSize := params.Get(pageSizeParam); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			pgSize = int64(ps)
		}
	}

	result, err := ctrl.svc.List(ctx, adapters.ListSuppliesParams{
		Version:  params.Get(versionParam),
		Page:     pg,
		PageSize: pgSize,
	})
	if err != nil {
		logger.Of(ctx).Debug("Failed to list supplies in service layer",
			zap.String("operation", "list_supplies"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		return result, err
	}
	return result, nil
}

func (ctrl *supplyController) Update(ctx context.Context, req *http.Request) (adapters.SupplyResponse, error) {
	dto := adapters.UpdateSupply{}
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		logger.Of(ctx).Debug("Failed to bind supply creation payload",
			zap.String("operation", "update_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		return adapters.SupplyResponse{}, err
	}

	id, _ := ctx.Value(idParam).(string)
	resp, err := ctrl.svc.Update(ctx, id, dto)
	if err != nil {
		logger.Of(ctx).Debug("Supply update failed in service layer",
			zap.String("operation", "update_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		return adapters.SupplyResponse{}, err
	}

	return resp, nil
}

func (ctrl *supplyController) Delete(ctx context.Context, req *http.Request) error {
	id, _ := ctx.Value(idParam).(string)
	if err := ctrl.svc.Delete(ctx, id); err != nil {
		logger.Of(ctx).Debug("Supply deletion failed in service layer",
			zap.String("operation", "delete_supply"),
			zap.Error(err),
			zap.String("entity", "supply"),
		)
		return err
	}

	return nil
}
