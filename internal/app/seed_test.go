package app

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSeedDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestRunSeed_DisabledDoesNotTouchDatabase(t *testing.T) {
	t.Setenv("SEED_ENABLED", "false")
	db, mock := newSeedDB(t)

	RunSeed(context.Background(), db)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_SkipsWhenAlreadySeeded(t *testing.T) {
	t.Setenv("SEED_ENABLED", "true")
	db, mock := newSeedDB(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "work"`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	RunSeed(context.Background(), db)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_InsertsWorksAndSupplies(t *testing.T) {
	t.Setenv("SEED_ENABLED", "true")
	db, mock := newSeedDB(t)
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "work"`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	for i := 0; i < 5; i++ {
		mock.ExpectExec(`INSERT INTO "work"`).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	for i := 0; i < 5; i++ {
		mock.ExpectExec(`INSERT INTO supply`).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	RunSeed(context.Background(), db)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_RollsBackOnInsertError(t *testing.T) {
	t.Setenv("SEED_ENABLED", "true")
	db, mock := newSeedDB(t)
	mock.ExpectQuery(`SELECT COUNT`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "work"`).WillReturnError(assertErr{})
	mock.ExpectRollback()

	RunSeed(context.Background(), db)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRunSeed_CountError(t *testing.T) {
	t.Setenv("SEED_ENABLED", "true")
	db, mock := newSeedDB(t)
	mock.ExpectQuery(`SELECT COUNT`).WillReturnError(assertErr{})

	RunSeed(context.Background(), db)

	assert.NoError(t, mock.ExpectationsWereMet())
}

type assertErr struct{}

func (assertErr) Error() string { return "boom" }
