package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/logger"
)

const checkSeededQuery = `SELECT COUNT(*) FROM "work"`

// RunSeed inserts the catalog demo data (works and supplies) when SEED_ENABLED=true.
// Seed rows use fixed IDs (see docs/fase4 resources contract) so e2e fixtures can rely on them.
// Idempotency is guaranteed by checking for existing rows in the work table.
func RunSeed(ctx context.Context, db *sql.DB) {
	log := logger.Global()

	if env.GetString("SEED_ENABLED", "false") != "true" {
		log.Info("seed: SEED_ENABLED is not true, skipping")
		return
	}

	seeded, err := isAlreadySeeded(ctx, db)
	if err != nil {
		log.Error(fmt.Errorf("seed: failed to check seed state: %w", err))
		return
	}
	if seeded {
		log.Info("seed: data already present, skipping")
		return
	}

	log.Info("seed: starting database seeding")

	if err := runInTransaction(ctx, db, insertSeedData); err != nil {
		log.Error(fmt.Errorf("seed: failed to insert seed data: %w", err))
		return
	}

	log.Info("seed: database seeded successfully")
}

func isAlreadySeeded(ctx context.Context, db *sql.DB) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, checkSeededQuery).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func runInTransaction(ctx context.Context, db *sql.DB, fn func(context.Context, *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func insertSeedData(ctx context.Context, tx *sql.Tx) error {
	if err := insertWorks(ctx, tx); err != nil {
		return err
	}
	logger.Global().Info("seed: works inserted")

	if err := insertSupplies(ctx, tx); err != nil {
		return err
	}
	logger.Global().Info("seed: supplies inserted")

	return nil
}

type workRow struct {
	id          string
	name        string
	description string
	unitPrice   string
}

func insertWorks(ctx context.Context, tx *sql.Tx) error {
	works := []workRow{
		{"11111111-1111-4111-8111-000000000001", "Troca de Óleo", "Troca de óleo do motor com filtro incluído", "120.00"},
		{"11111111-1111-4111-8111-000000000002", "Rodízio de Pneus", "Rodízio completo dos quatro pneus", "80.00"},
		{"11111111-1111-4111-8111-000000000003", "Alinhamento e Balanceamento", "Alinhamento de direção e balanceamento de rodas", "150.00"},
		{"11111111-1111-4111-8111-000000000004", "Revisão de Freios", "Inspeção e ajuste do sistema de freios", "200.00"},
		{"11111111-1111-4111-8111-000000000005", "Diagnóstico Eletrônico", "Leitura de falhas via scanner OBD-II", "100.00"},
	}

	const q = `INSERT INTO "work" (id, name, description, unit_price) VALUES ($1, $2, $3, $4)`

	for _, w := range works {
		if _, err := tx.ExecContext(ctx, q, w.id, w.name, w.description, w.unitPrice); err != nil {
			return fmt.Errorf("insert work %q: %w", w.name, err)
		}
	}
	return nil
}

type supplyRow struct {
	id          string
	name        string
	description string
	unitPrice   string
	quantity    int
}

func insertSupplies(ctx context.Context, tx *sql.Tx) error {
	supplies := []supplyRow{
		{"22222222-2222-4222-8222-000000000001", "Filtro de Óleo", "Filtro de óleo para motores 1.0 a 2.0", "35.00", 50},
		{"22222222-2222-4222-8222-000000000002", "Pastilha de Freio Dianteira", "Jogo de pastilhas de freio dianteiras", "120.00", 30},
		{"22222222-2222-4222-8222-000000000003", "Filtro de Ar", "Filtro de ar do motor", "45.00", 40},
		{"22222222-2222-4222-8222-000000000004", "Fluido de Freio DOT 4", "Fluido de freio DOT 4 - 500ml", "28.00", 60},
		{"22222222-2222-4222-8222-000000000005", "Correia Dentada", "Kit correia dentada com tensor e rolamento", "380.00", 15},
	}

	const q = `INSERT INTO supply (id, name, description, unit_price, stock_quantity) VALUES ($1, $2, $3, $4, $5)`

	for _, s := range supplies {
		if _, err := tx.ExecContext(ctx, q, s.id, s.name, s.description, s.unitPrice, s.quantity); err != nil {
			return fmt.Errorf("insert supply %q: %w", s.name, err)
		}
	}
	return nil
}
