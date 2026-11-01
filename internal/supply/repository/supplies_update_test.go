package repository_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func updateSupplyFixture() *domain.Supply {
	return domain.NewSupply("supply-1", "Brake Pad", "High performance brake pad", decimal.NewFromFloat(49.99), 10)
}

func TestSupplyRepository_UpdateWithVersion(t *testing.T) {
	updatePattern := regexp.QuoteMeta(`UPDATE "supply"`) + `[\s\S]*WHERE id = \$1 AND version = \$6`
	existsPattern := regexp.QuoteMeta(`SELECT EXISTS (SELECT 1 FROM "supply" WHERE id = $1)`)

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
		wantVer   int
	}{
		{
			name: "matching version updates and returns the new version",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(updatePattern).
					WithArgs("supply-1", "Brake Pad", "High performance brake pad", sqlmock.AnyArg(), 10, 3).
					WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(4))
				mock.ExpectCommit()
			},
			wantVer: 4,
		},
		{
			name: "stale version on an existing supply is a conflict",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(updatePattern).WillReturnRows(sqlmock.NewRows([]string{"version"}))
				mock.ExpectQuery(existsPattern).WithArgs("supply-1").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectRollback()
			},
			wantErr: domain.ErrSupplyVersionConflict,
		},
		{
			name: "unknown supply is not found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(updatePattern).WillReturnRows(sqlmock.NewRows([]string{"version"}))
				mock.ExpectQuery(existsPattern).WithArgs("supply-1").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
				mock.ExpectRollback()
			},
			wantErr: domain.ErrSupplyNotFound,
		},
		{
			name: "database error on update is returned",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(updatePattern).WillReturnError(errors.New("connection reset"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("connection reset"),
		},
		{
			name: "database error on existence check is returned",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(updatePattern).WillReturnRows(sqlmock.NewRows([]string{"version"}))
				mock.ExpectQuery(existsPattern).WillReturnError(errors.New("boom"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() { _ = db.Close() }()
			tt.setupMock(mock)

			supply := updateSupplyFixture()
			err = postgresdb.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
				return repository.NewPostgres().UpdateWithVersion(ctx, supply, 3)
			})

			if tt.wantErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.wantVer, supply.Version)
			} else {
				require.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrSupplyNotFound) || errors.Is(tt.wantErr, domain.ErrSupplyVersionConflict) {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
