package reservation

import (
	"context"
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/repository"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	soID     = "01JD8Z4R3Q8M2V7K5N9XWABCDE"
	supplyID = "22222222-2222-4222-8222-000000000001"
)

func newTestService(t *testing.T, stock int) (*reservationService, *repository.InMemory) {
	t.Helper()
	repo := repository.NewInMemory()
	repo.AddSupply(supplyID, "Filtro de Óleo", decimal.RequireFromString("35.5"), stock)
	return NewService(uow.NewInMemoryUoW(), repo, repo), repo
}

func TestReserve_DecrementsStockAndReturnsSnapshot(t *testing.T) {
	svc, repo := newTestService(t, 5)

	resp, err := svc.Reserve(context.Background(), adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 2})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ReservationID)
	assert.Equal(t, supplyID, resp.SupplyID)
	assert.Equal(t, "Filtro de Óleo", resp.Name)
	assert.Equal(t, "35.50", resp.UnitPrice)
	assert.Equal(t, 3, repo.Stock(supplyID))
}

func TestReserve_ValidationErrors(t *testing.T) {
	svc, _ := newTestService(t, 5)
	tests := []struct {
		name string
		req  adapters.ReserveRequest
		want error
	}{
		{"zero quantity", adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 0}, domain.ErrInvalidReservationQuantity},
		{"negative quantity", adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: -1}, domain.ErrInvalidReservationQuantity},
		{"empty service order", adapters.ReserveRequest{SupplyID: supplyID, Quantity: 1}, domain.ErrInvalidServiceOrderID},
		{"empty supply", adapters.ReserveRequest{ServiceOrderID: soID, Quantity: 1}, supplyDomain.ErrInvalidSupplyID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Reserve(context.Background(), tt.req)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}

func TestReserve_UnknownSupplyIsNotFound(t *testing.T) {
	svc, _ := newTestService(t, 5)
	_, err := svc.Reserve(context.Background(), adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: "missing", Quantity: 1})
	assert.ErrorIs(t, err, supplyDomain.ErrSupplyNotFound)
}

func TestReserve_InsufficientStockLeavesNoReservation(t *testing.T) {
	svc, repo := newTestService(t, 1)

	_, err := svc.Reserve(context.Background(), adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 2})

	assert.ErrorIs(t, err, domain.ErrInsufficientStock)
	assert.Equal(t, 1, repo.Stock(supplyID))
	released, _ := repo.ReleaseAllByServiceOrder(context.Background(), soID)
	assert.Empty(t, released, "no reservation must be left behind")
}

func TestReserve_SecondActiveReservationConflicts(t *testing.T) {
	svc, _ := newTestService(t, 10)
	ctx := context.Background()
	_, err := svc.Reserve(ctx, adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 1})
	require.NoError(t, err)

	_, err = svc.Reserve(ctx, adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 1})

	assert.ErrorIs(t, err, domain.ErrReservationAlreadyExists)
}

func TestReserve_ReactivatesAfterRelease(t *testing.T) {
	svc, repo := newTestService(t, 10)
	ctx := context.Background()
	first, err := svc.Reserve(ctx, adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 2})
	require.NoError(t, err)
	require.NoError(t, svc.Release(ctx, first.ReservationID))

	second, err := svc.Reserve(ctx, adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 3})

	require.NoError(t, err)
	assert.Equal(t, first.ReservationID, second.ReservationID, "upsert keeps the same reservation id")
	assert.Equal(t, 7, repo.Stock(supplyID))
}

func TestRelease_ReturnsStockOnceAndIsIdempotent(t *testing.T) {
	svc, repo := newTestService(t, 5)
	ctx := context.Background()
	res, err := svc.Reserve(ctx, adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 2})
	require.NoError(t, err)

	require.NoError(t, svc.Release(ctx, res.ReservationID))
	require.NoError(t, svc.Release(ctx, res.ReservationID))

	assert.Equal(t, 5, repo.Stock(supplyID))
}

func TestRelease_UnknownIsNotFound(t *testing.T) {
	svc, _ := newTestService(t, 5)
	assert.ErrorIs(t, svc.Release(context.Background(), "nope"), domain.ErrReservationNotFound)
	assert.ErrorIs(t, svc.Release(context.Background(), ""), domain.ErrReservationNotFound)
}

func TestReleaseAllByServiceOrder(t *testing.T) {
	svc, repo := newTestService(t, 10)
	ctx := context.Background()
	_, err := svc.Reserve(ctx, adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 2})
	require.NoError(t, err)
	_, err = svc.Reserve(ctx, adapters.ReserveRequest{ServiceOrderID: "OTHER", SupplyID: supplyID, Quantity: 3})
	require.NoError(t, err)

	require.NoError(t, svc.ReleaseAllByServiceOrder(ctx, soID))
	require.NoError(t, svc.ReleaseAllByServiceOrder(ctx, soID))

	assert.Equal(t, 7, repo.Stock(supplyID), "only the order's reservation (2 units) is released, once")
}

func TestReleaseAllByServiceOrder_EmptyIDIsRejected(t *testing.T) {
	svc, _ := newTestService(t, 1)
	assert.ErrorIs(t, svc.ReleaseAllByServiceOrder(context.Background(), ""), domain.ErrInvalidServiceOrderID)
}

type failingStock struct{ *repository.InMemory }

func (failingStock) Increment(context.Context, string, int) error { return errors.New("boom") }

func TestReleaseAllByServiceOrder_PropagatesStockError(t *testing.T) {
	repo := repository.NewInMemory()
	repo.AddSupply(supplyID, "x", decimal.NewFromInt(1), 5)
	svc := NewService(uow.NewInMemoryUoW(), repo, failingStock{repo})
	ctx := context.Background()
	_, err := svc.Reserve(ctx, adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 1})
	require.NoError(t, err)

	assert.Error(t, svc.ReleaseAllByServiceOrder(ctx, soID))
}

func TestReleaseStockHandler_ReleasesServiceOrder(t *testing.T) {
	svc, repo := newTestService(t, 4)
	ctx := context.Background()
	_, err := svc.Reserve(ctx, adapters.ReserveRequest{ServiceOrderID: soID, SupplyID: supplyID, Quantity: 4})
	require.NoError(t, err)

	h := ReleaseStockHandler(svc)
	require.NoError(t, h(ctx, messagingEnvelope(soID)))

	assert.Equal(t, 4, repo.Stock(supplyID))
}
