package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/db"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/logger"
	postgres "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"go.uber.org/zap"
)

const (
	upsertQuery = `
    INSERT INTO "supply" (id, name, description, unit_price, stock_quantity, version)
    VALUES ($1, $2, $3, $4, $5, $6)
    ON CONFLICT (id) DO UPDATE
    SET
        name           = EXCLUDED.name,
        description    = EXCLUDED.description,
        unit_price     = EXCLUDED.unit_price,
        stock_quantity = EXCLUDED.stock_quantity,
        version        = EXCLUDED.version,
        updated_at     = NOW()
`
	countSuppliesQuery = `SELECT COUNT(s.id) FROM "supply" s`
	searchQuery        = `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s`
	deleteQuery        = `DELETE FROM "supply" s where s.id = $1`
	// Optimistic lock: the row is updated only if its version is still the one the caller read.
	updateWithVersionQuery = `
    UPDATE "supply"
    SET name = $2, description = $3, unit_price = $4, stock_quantity = $5,
        version = version + 1, updated_at = NOW()
    WHERE id = $1 AND version = $6
    RETURNING version
`
	supplyExistsQuery = `SELECT EXISTS (SELECT 1 FROM "supply" WHERE id = $1)`
)

type postgresRepository struct{}

func NewPostgres() *postgresRepository {
	return &postgresRepository{}
}

func (r *postgresRepository) Save(ctx context.Context, supply *domain.Supply) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", upsertQuery), zap.Any("params", supply))
	if _, err = tx.ExecContext(
		ctx,
		upsertQuery,
		supply.ID,
		supply.Name,
		supply.Description,
		supply.UnitPrice,
		supply.StockQuantity,
		supply.Version,
	); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *postgresRepository) UpdateWithVersion(ctx context.Context, supply *domain.Supply, expectedVersion int) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", updateWithVersionQuery), zap.Any("params", supply.ID))
	var version int
	err = tx.QueryRowContext(ctx, updateWithVersionQuery,
		supply.ID,
		supply.Name,
		supply.Description,
		supply.UnitPrice,
		supply.StockQuantity,
		expectedVersion,
	).Scan(&version)
	if err == nil {
		supply.Version = version
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return pgPkg.Error(ctx, err)
	}

	// No row matched the id and version: tell "missing" apart from "stale".
	var exists bool
	if err := tx.QueryRowContext(ctx, supplyExistsQuery, supply.ID).Scan(&exists); err != nil {
		return pgPkg.Error(ctx, err)
	}
	if !exists {
		return domain.ErrSupplyNotFound
	}
	return domain.ErrSupplyVersionConflict
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", deleteQuery), zap.Any("params", id))
	if _, err = tx.ExecContext(ctx, deleteQuery, id); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *postgresRepository) Search(ctx context.Context, params adapters.ListSuppliesParams) ([]domain.Supply, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.PageSize
	offset := (params.Page - 1) * params.PageSize
	queryBuilder := db.QueryBuilder(searchQuery).
		OrderBy("s.created_at", db.ASC).
		AddPagination(limit, offset)
	if params.ID != "" {
		queryBuilder.Add("s.id = ", params.ID)
	}
	if params.Version != "" {
		queryBuilder.Add("s.version = ", params.Version)
	}
	query, args := queryBuilder.Build()

	logger.Of(ctx).Debug("Executing query", zap.String("query", query), zap.Any("params", args))
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	supplies := make([]domain.Supply, 0, limit)
	for rows.Next() {
		var supply domain.Supply
		if err = rows.Scan(
			&supply.ID,
			&supply.Name,
			&supply.Description,
			&supply.UnitPrice,
			&supply.StockQuantity,
			&supply.Version,
		); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		supplies = append(supplies, supply)
	}

	if err = rows.Err(); err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	return supplies, nil
}

func (r *postgresRepository) Count(ctx context.Context, params adapters.ListSuppliesParams) (int64, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return 0, err
	}

	queryBuilder := db.QueryBuilder(countSuppliesQuery)
	if params.ID != "" {
		queryBuilder.Add("s.id = ", params.ID)
	}
	if params.Version != "" {
		queryBuilder.Add("s.version = ", params.Version)
	}
	query, args := queryBuilder.Build()

	var total int64
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, pgPkg.Error(ctx, err)
	}

	return total, nil
}
