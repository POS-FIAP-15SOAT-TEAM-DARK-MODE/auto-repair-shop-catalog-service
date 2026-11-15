package messaging_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/messaging"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresDedupe_MarkProcessed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	d := messaging.NewPostgresDedupe()
	const id = "b7a1c6f0-2f64-4f6e-9a49-0d5a0b1e6c11"

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO processed_message`).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO processed_message`).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = uow.NewTransactionalUoW(db).Execute(context.Background(), func(ctx context.Context) error {
		first, err := d.MarkProcessed(ctx, id)
		require.NoError(t, err)
		second, err := d.MarkProcessed(ctx, id)
		require.NoError(t, err)
		assert.True(t, first)
		assert.False(t, second)
		return nil
	})
	require.NoError(t, err)
}

func TestPostgresDedupe_RequiresTransaction(t *testing.T) {
	_, err := messaging.NewPostgresDedupe().MarkProcessed(context.Background(), "id")
	assert.Error(t, err)
}

func TestNewSQSClient(t *testing.T) {
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ENDPOINT_URL", "http://localhost:4566")
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	client, err := messaging.NewSQSClient(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, client)
}
