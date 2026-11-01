package reservation

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/messaging"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/interfaces"
)

// ReleaseStockHandler consumes ReleaseStock commands (saga compensation): it
// releases every reservation of the command's service order.
func ReleaseStockHandler(svc interfaces.ReservationService) messaging.Handler {
	return func(ctx context.Context, env messaging.Envelope) error {
		return svc.ReleaseAllByServiceOrder(ctx, env.ServiceOrderID)
	}
}
