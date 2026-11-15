package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app"
	uowPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWorkDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	app.ConnectWithDB(db) // reads use the shared handle, as in production
	return db, mock
}

func TestWorkPostgres_SaveInsideUnitOfWork(t *testing.T) {
	db, mock := newWorkDB(t)
	w := &domain.Work{ID: "w1", Name: "Troca", Description: "Troca de óleo", Price: decimal.RequireFromString("150"), Status: domain.ACTIVE}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "work"`).WithArgs("w1", "Troca", "Troca de óleo", w.Price, true).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := uowPkg.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		return repository.NewPostgres().Save(ctx, w)
	})
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWorkPostgres_SaveWithoutTransactionFails(t *testing.T) {
	newWorkDB(t)
	err := repository.NewPostgres().Save(context.Background(), &domain.Work{})
	assert.ErrorIs(t, err, uowPkg.ErrMissingPostgresTransaction)
}

func TestWorkPostgres_FindByID(t *testing.T) {
	_, mock := newWorkDB(t)
	repo := repository.NewPostgres()
	cols := []string{"id", "name", "description", "unit_price", "status"}

	mock.ExpectQuery(`SELECT s.id, s.name`).WithArgs("w1").
		WillReturnRows(sqlmock.NewRows(cols).AddRow("w1", "Troca", "desc", "150.00", true))
	w, err := repo.FindByID(context.Background(), "w1")
	require.NoError(t, err)
	assert.Equal(t, "Troca", w.Name)
	assert.True(t, w.Status.Bool())

	mock.ExpectQuery(`SELECT s.id, s.name`).WithArgs("missing").WillReturnError(sql.ErrNoRows)
	_, err = repo.FindByID(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrWorkNotFound)

	mock.ExpectQuery(`SELECT s.id, s.name`).WithArgs("broken").WillReturnError(errors.New("io"))
	_, err = repo.FindByID(context.Background(), "broken")
	assert.Error(t, err)
}

func TestWorkPostgres_CountAndSearchWithStatusFilter(t *testing.T) {
	_, mock := newWorkDB(t)
	repo := repository.NewPostgres()
	ctx := context.Background()
	params := adapters.ListWorksParams{Page: 2, PageSize: 5, Status: "ACTIVE"}

	mock.ExpectQuery(`SELECT COUNT\(s.id\) FROM "work" s WHERE status = \$1`).WithArgs(true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))
	total, err := repo.Count(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, int64(7), total)

	cols := []string{"id", "name", "description", "unit_price", "status"}
	mock.ExpectQuery(`SELECT s.id, s.name, s.description, s.unit_price, s.status FROM "work" s WHERE status = \$1`).
		WillReturnRows(sqlmock.NewRows(cols).AddRow("w1", "Troca", "desc", "150.00", true).AddRow("w2", "Freio", "desc", "200.00", false))
	works, err := repo.Search(ctx, params)
	require.NoError(t, err)
	assert.Len(t, works, 2)

	mock.ExpectQuery(`SELECT COUNT`).WillReturnError(errors.New("boom"))
	_, err = repo.Count(ctx, adapters.ListWorksParams{})
	assert.Error(t, err)
}

func TestWorkPostgres_Delete(t *testing.T) {
	db, mock := newWorkDB(t)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "work"`).WithArgs("w1").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := uowPkg.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		return repository.NewPostgres().Delete(ctx, "w1")
	})
	require.NoError(t, err)
}
