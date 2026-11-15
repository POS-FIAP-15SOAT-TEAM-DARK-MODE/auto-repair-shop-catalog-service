package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestSupplyRepository_Save(t *testing.T) {
	upsertPattern := `INSERT INTO "supply" \(id, name, description, unit_price, stock_quantity, version\)\s*VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock, s *domain.Supply)
		supply    *domain.Supply
		useTx     bool
		wantErr   error
	}{
		{
			name: "upsert insert success",
			setupMock: func(mock sqlmock.Sqlmock, v *domain.Supply) {
				mock.ExpectBegin()
				mock.ExpectExec(upsertPattern).
					WithArgs(v.ID, v.Name, v.Description, v.UnitPrice, v.StockQuantity, v.Version).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			supply: &domain.Supply{
				ID:            "supply-1",
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
			useTx: true,
		},
		{
			name: "upsert update success",
			setupMock: func(mock sqlmock.Sqlmock, v *domain.Supply) {
				mock.ExpectBegin()
				mock.ExpectExec(upsertPattern).
					WithArgs(v.ID, v.Name, v.Description, v.UnitPrice, v.StockQuantity, v.Version).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			supply: &domain.Supply{
				ID:            "supply-1",
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
			useTx: true,
		},
		{
			name: "upsert exec error",
			setupMock: func(mock sqlmock.Sqlmock, v *domain.Supply) {
				mock.ExpectBegin()
				mock.ExpectExec(upsertPattern).
					WithArgs(v.ID, v.Name, v.Description, v.UnitPrice, v.StockQuantity, v.Version).
					WillReturnError(errors.New("exec error"))
				mock.ExpectRollback()
			},
			supply: &domain.Supply{
				ID:            "supply-1",
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
			useTx:   true,
			wantErr: errors.New("exec error"),
		},
		{
			name:    "missing transaction",
			supply:  &domain.Supply{},
			useTx:   false,
			wantErr: postgresdb.ErrMissingPostgresTransaction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewPostgres()

			if tt.useTx {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				defer func() { _ = db.Close() }()

				uow := postgresdb.NewTransactionalUoW(db)
				if tt.setupMock != nil {
					tt.setupMock(mock, tt.supply)
				}

				err = uow.Execute(context.Background(), func(ctx context.Context) error {
					return repo.Save(ctx, tt.supply)
				})
				if tt.wantErr != nil {
					assert.Error(t, err)
					if errors.Is(tt.wantErr, domain.ErrSupplyNotFound) || errors.Is(tt.wantErr, postgresdb.ErrMissingPostgresTransaction) {
						assert.ErrorIs(t, err, tt.wantErr)
					} else {
						assert.Contains(t, err.Error(), tt.wantErr.Error())
					}
					return
				}
				assert.NoError(t, err)
				return
			}

			err := repo.Save(context.Background(), tt.supply)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSupplyRepository_Delete(t *testing.T) {
	deletePattern := `DELETE FROM "supply" s where s.id = \$1`

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock, id string)
		id        string
		useTx     bool
		wantErr   error
	}{
		{
			name:    "delete missing tx",
			id:      "id-1",
			useTx:   false,
			wantErr: postgresdb.ErrMissingPostgresTransaction,
		},
		{
			name: "delete success",
			setupMock: func(mock sqlmock.Sqlmock, id string) {
				mock.ExpectBegin()
				mock.ExpectExec(deletePattern).
					WithArgs(id).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			id:    "id-1",
			useTx: true,
		},
		{
			name: "delete exec error",
			setupMock: func(mock sqlmock.Sqlmock, id string) {
				mock.ExpectBegin()
				mock.ExpectExec(deletePattern).
					WithArgs(id).
					WillReturnError(errors.New("exec error"))
				mock.ExpectRollback()
			},
			id:      "id-1",
			useTx:   true,
			wantErr: errors.New("exec error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewPostgres()

			if tt.useTx {
				db, mock, err := sqlmock.New()
				assert.NoError(t, err)
				defer func() { _ = db.Close() }()

				uow := postgresdb.NewTransactionalUoW(db)
				if tt.setupMock != nil {
					tt.setupMock(mock, tt.id)
				}

				err = uow.Execute(context.Background(), func(ctx context.Context) error {
					return repo.Delete(ctx, tt.id)
				})
				if tt.wantErr != nil {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.wantErr.Error())
					return
				}
				assert.NoError(t, err)
				return
			}

			err := repo.Delete(context.Background(), tt.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSupplyRepository_Search(t *testing.T) {
	supplyColumns := []string{"id", "name", "description", "unit_price", "stock_quantity", "version"}

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		params    adapters.ListSuppliesParams
		wantLen   int
		wantErr   error
		wantFirst *domain.Supply
	}{
		{
			name: "search success no filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s ORDER BY s.created_at ASC LIMIT $1`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(int64(10)).
					WillReturnRows(sqlmock.NewRows(supplyColumns).
						AddRow("supply-1", "Brake Pad", "High performance brake pad", decimal.NewFromFloat(49.99), 10, 1))
			},
			params: adapters.ListSuppliesParams{
				Page:     1,
				PageSize: 10,
			},
			wantLen: 1,
			wantFirst: &domain.Supply{
				ID:            "supply-1",
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
		},
		{
			name: "search with version filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s WHERE s.version = $1 ORDER BY s.created_at ASC LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("1", int64(10)).
					WillReturnRows(sqlmock.NewRows(supplyColumns).
						AddRow("supply-1", "Brake Pad", "High performance brake pad", decimal.NewFromFloat(49.99), 10, 1))
			},
			params: adapters.ListSuppliesParams{
				Version:  "1",
				Page:     1,
				PageSize: 10,
			},
			wantFirst: &domain.Supply{
				ID:            "supply-1",
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
			wantLen: 1,
		},
		{
			name: "search with id filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s WHERE s.id = $1 ORDER BY s.created_at ASC LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("supply-1", int64(10)).
					WillReturnRows(sqlmock.NewRows(supplyColumns).
						AddRow("supply-1", "Brake Pad", "High performance brake pad", decimal.NewFromFloat(49.99), 10, 1))
			},
			params: adapters.ListSuppliesParams{
				ID:       "supply-1",
				Page:     1,
				PageSize: 10,
			},
			wantFirst: &domain.Supply{
				ID:            "supply-1",
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
			wantLen: 1,
		},
		{
			name: "search with all filters",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s WHERE s.id = $1 AND s.version = $2 ORDER BY s.created_at ASC LIMIT $3`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("supply-1", "1", int64(10)).
					WillReturnRows(sqlmock.NewRows(supplyColumns).
						AddRow("supply-1", "Brake Pad", "High performance brake pad", decimal.NewFromFloat(49.99), 10, 1))
			},
			params: adapters.ListSuppliesParams{
				ID:       "supply-1",
				Version:  "1",
				Page:     1,
				PageSize: 10,
			},
			wantFirst: &domain.Supply{
				ID:            "supply-1",
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
			wantLen: 1,
		},
		{
			name: "search empty result",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s WHERE s.version = $1 ORDER BY s.created_at ASC LIMIT $2`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("NOTFOUND", int64(10)).
					WillReturnRows(sqlmock.NewRows(supplyColumns))
			},
			params: adapters.ListSuppliesParams{
				Version:  "NOTFOUND",
				Page:     1,
				PageSize: 10,
			},
			wantLen: 0,
		},
		{
			name: "search query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s ORDER BY s.created_at ASC LIMIT $1`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(int64(10)).
					WillReturnError(errors.New("db error"))
			},
			params: adapters.ListSuppliesParams{
				Page:     1,
				PageSize: 10,
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "search scan error",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s LIMIT $1`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(int64(10)).
					WillReturnRows(sqlmock.NewRows(supplyColumns).
						AddRow("supply-1", "Brake Pad", "High performance brake pad", decimal.NewFromFloat(49.99), 10, 1))
			},
			params: adapters.ListSuppliesParams{
				Page:     1,
				PageSize: 10,
			},
			wantErr: errors.New("sql"),
		},
		{
			name: "search rows error",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s ORDER BY s.created_at ASC LIMIT $1`
				rows := sqlmock.NewRows(supplyColumns).
					AddRow("supply-1", "Brake Pad", "High performance brake pad", decimal.NewFromFloat(49.99), 10, 1)
				rows.RowError(0, errors.New("row error"))
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(int64(10)).
					WillReturnRows(rows)
			},
			params: adapters.ListSuppliesParams{
				Page:     1,
				PageSize: 10,
			},
			wantErr: errors.New("row error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer func() { _ = db.Close() }()

			app.ConnectWithDB(db)
			repo := repository.NewPostgres()

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			got, err := repo.Search(context.Background(), tt.params)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
			if tt.wantFirst != nil && len(got) > 0 {
				assert.Equal(t, *tt.wantFirst, got[0])
			}
		})
	}
}

func TestSupplyRepository_Count(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		params    adapters.ListSuppliesParams
		wantCount int64
		wantErr   error
	}{
		{
			name: "count success no filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(s.id) FROM "supply" s`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			wantCount: 10,
		},
		{
			name: "count with version filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(s.id) FROM "supply" s WHERE s.version =`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			params: adapters.ListSuppliesParams{
				Version: "1",
			},
			wantCount: 10,
		},
		{
			name: "count with id filter",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(s.id) FROM "supply" s WHERE s.id =`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("vh-1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			params: adapters.ListSuppliesParams{
				ID: "vh-1",
			},
			wantCount: 10,
		},
		{
			name: "count with all filters",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(s.id) FROM "supply" s WHERE s.id = $1 AND s.version = $2`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs("vh-1", "1").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
			},
			params: adapters.ListSuppliesParams{
				ID:      "vh-1",
				Version: "1",
			},
			wantCount: 10,
		},
		{
			name: "count query error",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(s.id) FROM "supply" s`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "count scan error",
			setupMock: func(mock sqlmock.Sqlmock) {
				query := `SELECT COUNT(s.id) FROM "supply" s`
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(sqlmock.NewRows([]string{"invalid"}))
				mock.ExpectRollback()
			},
			wantErr: sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer func() { _ = db.Close() }()

			app.ConnectWithDB(db)
			repo := repository.NewPostgres()

			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			count, err := repo.Count(context.Background(), tt.params)
			assert.Equal(t, tt.wantCount, count)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
