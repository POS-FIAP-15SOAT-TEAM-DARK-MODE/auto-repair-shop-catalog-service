package reservation

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/interfaces"
)

type reservationService struct {
	uow          uow.Executor
	reservations interfaces.ReservationRepository
	stock        interfaces.StockRepository
}

func NewService(uow uow.Executor, reservations interfaces.ReservationRepository, stock interfaces.StockRepository) *reservationService {
	return &reservationService{uow: uow, reservations: reservations, stock: stock}
}

// Reserve holds qty units of a supply for a service order. The reservation row
// and the stock decrement happen in the same unit of work.
func (s *reservationService) Reserve(ctx context.Context, req adapters.ReserveRequest) (adapters.ReserveResponse, error) {
	res, err := domain.NewReservation(req.ServiceOrderID, req.SupplyID, req.Quantity)
	if err != nil {
		return adapters.ReserveResponse{}, err
	}

	var supply domain.SupplyInfo
	var version int
	if err := s.uow.Execute(ctx, func(txCtx context.Context) error {
		var err error
		supply, err = s.stock.FindSupply(txCtx, res.SupplyID)
		if err != nil {
			return err
		}
		if supply.StockQuantity < res.Quantity {
			return domain.ErrInsufficientStock
		}
		if err := s.reservations.Reserve(txCtx, res); err != nil {
			return err
		}
		version, err = s.stock.Decrement(txCtx, res.SupplyID, res.Quantity)
		return err
	}); err != nil {
		return adapters.ReserveResponse{}, err
	}

	return adapters.ReserveResponse{
		ReservationID: res.ID,
		SupplyID:      supply.ID,
		Name:          supply.Name,
		Description:   supply.Description,
		UnitPrice:     supply.UnitPrice.StringFixed(2),
		Version:       version,
	}, nil
}

// Release returns the stock of one reservation. A reservation that is already
// RELEASED is a no-op; an unknown id yields domain.ErrReservationNotFound.
func (s *reservationService) Release(ctx context.Context, reservationID string) error {
	if reservationID == "" {
		return domain.ErrReservationNotFound
	}

	return s.uow.Execute(ctx, func(txCtx context.Context) error {
		released, err := s.reservations.Release(txCtx, reservationID)
		if err != nil {
			return err
		}
		if released == nil {
			// Already released (no-op) or unknown (404): FindByID tells them apart.
			_, err := s.reservations.FindByID(txCtx, reservationID)
			return err
		}
		return s.stock.Increment(txCtx, released.SupplyID, released.Quantity)
	})
}

// ReleaseAllByServiceOrder returns the stock of every RESERVED reservation of a
// service order. Calling it again finds nothing to release and changes nothing.
func (s *reservationService) ReleaseAllByServiceOrder(ctx context.Context, serviceOrderID string) error {
	if serviceOrderID == "" {
		return domain.ErrInvalidServiceOrderID
	}

	return s.uow.Execute(ctx, func(txCtx context.Context) error {
		released, err := s.reservations.ReleaseAllByServiceOrder(txCtx, serviceOrderID)
		if err != nil {
			return err
		}
		for _, r := range released {
			if err := s.stock.Increment(txCtx, r.SupplyID, r.Quantity); err != nil {
				return err
			}
		}
		return nil
	})
}
