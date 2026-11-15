package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestPostgresReserve_Inserted(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPostgresReservations()
	res := &domain.Reservation{ID: "r-new", ServiceOrderID: "so", SupplyID: "s", Quantity: 2}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO stock_reservation")).
		WithArgs("r-new", "so", "s", 2).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("r-new"))
	mock.ExpectCommit()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		return repo.Reserve(ctx, res)
	})

	require.NoError(t, err)
	assert.Equal(t, "r-new", res.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresReserve_ReactivatedKeepsStoredID(t *testing.T) {
	db, mock := newMock(t)
	res := &domain.Reservation{ID: "fresh", ServiceOrderID: "so", SupplyID: "s", Quantity: 2}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO stock_reservation").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("stored"))
	mock.ExpectCommit()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		return NewPostgresReservations().Reserve(ctx, res)
	})

	require.NoError(t, err)
	assert.Equal(t, "stored", res.ID)
}

func TestPostgresReserve_ConflictWhenActive(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO stock_reservation").WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		return NewPostgresReservations().Reserve(ctx, &domain.Reservation{ID: "x"})
	})

	assert.ErrorIs(t, err, domain.ErrReservationAlreadyExists)
}

func TestPostgresReserve_DatabaseError(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO stock_reservation").WillReturnError(errors.New("connection reset"))
	mock.ExpectRollback()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		return NewPostgresReservations().Reserve(ctx, &domain.Reservation{ID: "x"})
	})

	assert.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrReservationAlreadyExists)
}

func TestPostgresFindByID(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPostgresReservations()
	cols := []string{"id", "service_order_id", "supply_id", "quantity", "status"}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, service_order_id").WithArgs("r1").
		WillReturnRows(sqlmock.NewRows(cols).AddRow("r1", "so", "s", 3, "RESERVED"))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, service_order_id").WithArgs("missing").WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		got, err := repo.FindByID(ctx, "r1")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusReserved, got.Status)
		assert.Equal(t, 3, got.Quantity)
		return nil
	})
	require.NoError(t, err)

	err = uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.FindByID(ctx, "missing")
		return err
	})
	assert.ErrorIs(t, err, domain.ErrReservationNotFound)
}

func TestPostgresRelease(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPostgresReservations()
	cols := []string{"id", "service_order_id", "supply_id", "quantity", "status"}

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE stock_reservation").WithArgs("r1").
		WillReturnRows(sqlmock.NewRows(cols).AddRow("r1", "so", "s", 2, "RELEASED"))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE stock_reservation").WithArgs("r2").WillReturnError(sql.ErrNoRows)
	mock.ExpectCommit()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		got, err := repo.Release(ctx, "r1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "s", got.SupplyID)
		return nil
	})
	require.NoError(t, err)

	err = uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		got, err := repo.Release(ctx, "r2")
		assert.NoError(t, err)
		assert.Nil(t, got)
		return nil
	})
	require.NoError(t, err)
}

func TestPostgresReleaseAllByServiceOrder(t *testing.T) {
	db, mock := newMock(t)
	cols := []string{"id", "service_order_id", "supply_id", "quantity", "status"}

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE stock_reservation").WithArgs("so").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("r1", "so", "s1", 2, "RELEASED").
			AddRow("r2", "so", "s2", 1, "RELEASED"))
	mock.ExpectCommit()

	var released []domain.Reservation
	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		var err error
		released, err = NewPostgresReservations().ReleaseAllByServiceOrder(ctx, "so")
		return err
	})

	require.NoError(t, err)
	assert.Len(t, released, 2)
}

func TestPostgresReleaseAllByServiceOrder_QueryError(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE stock_reservation").WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		_, err := NewPostgresReservations().ReleaseAllByServiceOrder(ctx, "so")
		return err
	})
	assert.Error(t, err)
}

func TestPostgresStock_FindSupply(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPostgresStock()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, name, COALESCE").WithArgs("s1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "unit_price", "stock", "version"}).AddRow("s1", "Filtro", "Filtro de óleo", "35.50", 4, 7))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, name, COALESCE").WithArgs("nope").WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		info, err := repo.FindSupply(ctx, "s1")
		require.NoError(t, err)
		assert.Equal(t, "35.5", info.UnitPrice.String())
		assert.Equal(t, 4, info.StockQuantity)
		assert.Equal(t, "Filtro de óleo", info.Description)
		assert.Equal(t, 7, info.Version)
		return nil
	})
	require.NoError(t, err)

	err = uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.FindSupply(ctx, "nope")
		return err
	})
	assert.ErrorIs(t, err, supplyDomain.ErrSupplyNotFound)
}

func TestPostgresStock_FindSupply_BadPrice(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, name, COALESCE").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "d", "p", "s", "v"}).AddRow("s", "n", "d", "abc", 1, 1))
	mock.ExpectRollback()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		_, err := NewPostgresStock().FindSupply(ctx, "s")
		return err
	})
	assert.Error(t, err)
}

func TestPostgresStock_DecrementAndIncrement(t *testing.T) {
	db, mock := newMock(t)
	repo := NewPostgresStock()

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE supply SET stock_quantity = stock_quantity -.*RETURNING version").WithArgs(2, "s1").
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(5))
	mock.ExpectQuery("UPDATE supply SET stock_quantity = stock_quantity -.*RETURNING version").WithArgs(9, "s1").
		WillReturnRows(sqlmock.NewRows([]string{"version"}))
	mock.ExpectExec("UPDATE supply SET stock_quantity = stock_quantity \\+").WithArgs(2, "s1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		version, err := repo.Decrement(ctx, "s1", 2)
		require.NoError(t, err)
		assert.Equal(t, 5, version)
		_, err = repo.Decrement(ctx, "s1", 9)
		assert.ErrorIs(t, err, domain.ErrInsufficientStock)
		return repo.Increment(ctx, "s1", 2)
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStock_ExecErrors(t *testing.T) {
	db, mock := newMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE supply").WillReturnError(errors.New("boom"))
	mock.ExpectExec("UPDATE supply").WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	err := uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		_, err := NewPostgresStock().Decrement(ctx, "s", 1)
		assert.Error(t, err)
		return NewPostgresStock().Increment(ctx, "s", 1)
	})
	assert.Error(t, err)
}
