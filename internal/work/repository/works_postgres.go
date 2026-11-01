package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/db"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/logger"
	uowPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
	"go.uber.org/zap"
)

const (
	upsertQuery = `
		INSERT INTO "work" (id, name, description, unit_price, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
		SET
			name        = EXCLUDED.name,
			description = EXCLUDED.description,
			unit_price  = EXCLUDED.unit_price,
			status      = EXCLUDED.status,
			updated_at  = NOW()
	`
	countQuery    = `SELECT COUNT(s.id) FROM "work" s`
	searchQuery   = `SELECT s.id, s.name, s.description, s.unit_price, s.status FROM "work" s`
	deleteQuery   = `DELETE FROM "work" s WHERE s.id = $1`
	findByIDQuery = `SELECT s.id, s.name, s.description, s.unit_price, s.status FROM "work" s WHERE s.id = $1`
)

type postgresRepository struct{}

func NewPostgres() *postgresRepository {
	return &postgresRepository{}
}

func (r *postgresRepository) Save(ctx context.Context, w *domain.Work) error {
	tx, err := uowPkg.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", upsertQuery), zap.Any("params", w))
	if _, err = tx.ExecContext(ctx, upsertQuery, w.ID, w.Name, w.Description, w.Price, w.Status.Bool()); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}

func (r *postgresRepository) FindByID(ctx context.Context, id string) (domain.Work, error) {
	tx, err := uowPkg.GetOneTimeTransaction(ctx)
	if err != nil {
		return domain.Work{}, err
	}

	var w domain.Work
	err = tx.QueryRowContext(ctx, findByIDQuery, id).Scan(&w.ID, &w.Name, &w.Description, &w.Price, &w.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Work{}, domain.ErrWorkNotFound
	}
	if err != nil {
		return domain.Work{}, pgPkg.Error(ctx, err)
	}

	return w, nil
}

func (r *postgresRepository) Count(ctx context.Context, params adapters.ListWorksParams) (int64, error) {
	tx, err := uowPkg.GetOneTimeTransaction(ctx)
	if err != nil {
		return 0, err
	}

	qb := db.QueryBuilder(countQuery)
	if params.Status != "" {
		qb.Add("status = ", domain.StringToWorkStatus(params.Status).Bool())
	}
	query, args := qb.Build()

	var total int64
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, pgPkg.Error(ctx, err)
	}

	return total, nil
}

func (r *postgresRepository) Search(ctx context.Context, params adapters.ListWorksParams) ([]domain.Work, error) {
	tx, err := uowPkg.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	limit := params.PageSize
	offset := (params.Page - 1) * params.PageSize
	qb := db.QueryBuilder(searchQuery).OrderBy("s.created_at", db.ASC).AddPagination(limit, offset)
	if params.Status != "" {
		qb.Add("status = ", domain.StringToWorkStatus(params.Status).Bool())
	}
	query, args := qb.Build()

	logger.Of(ctx).Debug("Executing query", zap.String("query", query), zap.Any("params", args))
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	works := make([]domain.Work, 0, limit)
	for rows.Next() {
		var w domain.Work
		if err = rows.Scan(&w.ID, &w.Name, &w.Description, &w.Price, &w.Status); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}
		works = append(works, w)
	}

	if err = rows.Err(); err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	return works, nil
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	tx, err := uowPkg.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", deleteQuery), zap.Any("params", id))
	if _, err = tx.ExecContext(ctx, deleteQuery, id); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}
