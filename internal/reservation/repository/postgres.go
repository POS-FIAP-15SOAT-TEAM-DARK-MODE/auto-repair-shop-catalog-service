package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/shopspring/decimal"
)

const (
	reserveQuery = `
		INSERT INTO stock_reservation (id, service_order_id, supply_id, quantity, status)
		VALUES ($1, $2, $3, $4, 'RESERVED')
		ON CONFLICT (service_order_id, supply_id) DO UPDATE
		SET status = 'RESERVED', quantity = EXCLUDED.quantity, released_at = NULL
		WHERE stock_reservation.status = 'RELEASED'
		RETURNING id
	`
	findReservationQuery = `
		SELECT id, service_order_id, supply_id, quantity, status
		FROM stock_reservation WHERE id = $1
	`
	releaseQuery = `
		UPDATE stock_reservation
		SET status = 'RELEASED', released_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status = 'RESERVED'
		RETURNING id, service_order_id, supply_id, quantity, status
	`
	releaseByServiceOrderQuery = `
		UPDATE stock_reservation
		SET status = 'RELEASED', released_at = CURRENT_TIMESTAMP
		WHERE service_order_id = $1 AND status = 'RESERVED'
		RETURNING id, service_order_id, supply_id, quantity, status
	`
	findSupplyQuery = `SELECT id, name, COALESCE(description, ''), unit_price, stock_quantity, version FROM supply WHERE id = $1`
	decrementQuery  = `
		UPDATE supply SET stock_quantity = stock_quantity - $1, version = version + 1
		WHERE id = $2 AND stock_quantity >= $1
		RETURNING version
	`
	incrementQuery = `
		UPDATE supply SET stock_quantity = stock_quantity + $1, version = version + 1
		WHERE id = $2
	`
)

// execer is satisfied by *sql.Tx (inside a unit of work) and *sql.DB (plain reads).
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func conn(ctx context.Context) (execer, error) {
	if tx, err := uow.GetTransaction(ctx); err == nil {
		return tx, nil
	}
	db, err := uow.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

type postgresReservationRepository struct{}

func NewPostgresReservations() *postgresReservationRepository {
	return &postgresReservationRepository{}
}

func (r *postgresReservationRepository) Reserve(ctx context.Context, res *domain.Reservation) error {
	db, err := conn(ctx)
	if err != nil {
		return err
	}

	var id string
	err = db.QueryRowContext(ctx, reserveQuery, res.ID, res.ServiceOrderID, res.SupplyID, res.Quantity).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrReservationAlreadyExists
	}
	if err != nil {
		return postgres.Error(ctx, err)
	}

	res.ID = id
	res.Status = domain.StatusReserved
	return nil
}

func (r *postgresReservationRepository) FindByID(ctx context.Context, id string) (domain.Reservation, error) {
	db, err := conn(ctx)
	if err != nil {
		return domain.Reservation{}, err
	}

	res, err := scanOne(db.QueryRowContext(ctx, findReservationQuery, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Reservation{}, domain.ErrReservationNotFound
	}
	if err != nil {
		return domain.Reservation{}, postgres.Error(ctx, err)
	}
	return res, nil
}

func (r *postgresReservationRepository) Release(ctx context.Context, id string) (*domain.Reservation, error) {
	db, err := conn(ctx)
	if err != nil {
		return nil, err
	}

	res, err := scanOne(db.QueryRowContext(ctx, releaseQuery, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, postgres.Error(ctx, err)
	}
	return &res, nil
}

func (r *postgresReservationRepository) ReleaseAllByServiceOrder(ctx context.Context, serviceOrderID string) ([]domain.Reservation, error) {
	db, err := conn(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, releaseByServiceOrderQuery, serviceOrderID)
	if err != nil {
		return nil, postgres.Error(ctx, err)
	}
	defer func() { _ = rows.Close() }()

	var released []domain.Reservation
	for rows.Next() {
		res, err := scanRow(rows)
		if err != nil {
			return nil, postgres.Error(ctx, err)
		}
		released = append(released, res)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.Error(ctx, err)
	}
	return released, nil
}

type postgresStockRepository struct{}

func NewPostgresStock() *postgresStockRepository {
	return &postgresStockRepository{}
}

func (r *postgresStockRepository) FindSupply(ctx context.Context, id string) (domain.SupplyInfo, error) {
	db, err := conn(ctx)
	if err != nil {
		return domain.SupplyInfo{}, err
	}

	var info domain.SupplyInfo
	var price string
	err = db.QueryRowContext(ctx, findSupplyQuery, id).Scan(&info.ID, &info.Name, &info.Description, &price, &info.StockQuantity, &info.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.SupplyInfo{}, supplyDomain.ErrSupplyNotFound
	}
	if err != nil {
		return domain.SupplyInfo{}, postgres.Error(ctx, err)
	}

	info.UnitPrice, err = decimal.NewFromString(price)
	if err != nil {
		return domain.SupplyInfo{}, err
	}
	return info, nil
}

func (r *postgresStockRepository) Decrement(ctx context.Context, supplyID string, qty int) (int, error) {
	db, err := conn(ctx)
	if err != nil {
		return 0, err
	}

	var version int
	err = db.QueryRowContext(ctx, decrementQuery, qty, supplyID).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrInsufficientStock
	}
	if err != nil {
		return 0, postgres.Error(ctx, err)
	}
	return version, nil
}

func (r *postgresStockRepository) Increment(ctx context.Context, supplyID string, qty int) error {
	db, err := conn(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, incrementQuery, qty, supplyID); err != nil {
		return postgres.Error(ctx, err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanOne(row scanner) (domain.Reservation, error) {
	return scanRow(row)
}

func scanRow(s scanner) (domain.Reservation, error) {
	var res domain.Reservation
	var status string
	if err := s.Scan(&res.ID, &res.ServiceOrderID, &res.SupplyID, &res.Quantity, &status); err != nil {
		return domain.Reservation{}, err
	}
	res.Status = domain.Status(status)
	return res, nil
}
