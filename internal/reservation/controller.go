package reservation

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/json"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/interfaces"
)

const idParam = "id"

type reservationController struct {
	svc interfaces.ReservationService
}

func NewController(svc interfaces.ReservationService) *reservationController {
	return &reservationController{svc: svc}
}

func (ctrl *reservationController) Reserve(ctx context.Context, req *http.Request) (adapters.ReserveResponse, error) {
	var dto adapters.ReserveRequest
	if err := json.ParseJsonBodyToStruct(req.Body, &dto); err != nil {
		return adapters.ReserveResponse{}, err
	}
	return ctrl.svc.Reserve(ctx, dto)
}

func (ctrl *reservationController) Release(ctx context.Context, _ *http.Request) error {
	id, _ := ctx.Value(idParam).(string)
	return ctrl.svc.Release(ctx, id)
}
