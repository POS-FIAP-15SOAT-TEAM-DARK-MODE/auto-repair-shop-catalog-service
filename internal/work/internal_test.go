package work

import (
	"context"
	"net/http"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternalService_FindByID(t *testing.T) {
	repo := repository.NewInMemory()
	w, err := domain.NewWork("Troca de óleo", "Troca de óleo do motor", "150.00", domain.ACTIVE)
	require.NoError(t, err)
	w.ID = "9b2d6a0e-3c0f-4a0b-8d57-2a5f6a1c9e11"
	w.Description = "Troca de óleo do motor com filtro incluído"
	require.NoError(t, repo.Save(context.Background(), w))
	svc := NewInternalService(uow.NewInMemoryUoW(), repo)

	got, err := svc.FindByID(context.Background(), w.ID)

	require.NoError(t, err)
	assert.Equal(t, w.ID, got.ID)
	assert.Equal(t, w.Description, got.Description)
	assert.Equal(t, "150.00", got.UnitPrice)
	assert.True(t, got.Active)
}

func TestInternalService_FindByID_Errors(t *testing.T) {
	svc := NewInternalService(uow.NewInMemoryUoW(), repository.NewInMemory())

	_, err := svc.FindByID(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)

	_, err = svc.FindByID(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrWorkNotFound)
}

func TestInternalController_Get(t *testing.T) {
	repo := repository.NewInMemory()
	w := &domain.Work{ID: "w1", Name: "Revisão", Price: decimal.RequireFromString("200"), Status: domain.INACTIVE}
	require.NoError(t, repo.Save(context.Background(), w))
	ctrl := NewInternalController(NewInternalService(uow.NewInMemoryUoW(), repo))

	ctx := context.WithValue(context.Background(), idParam, "w1") //nolint:staticcheck // route params are string keys by design
	got, err := ctrl.Get(ctx, &http.Request{})

	require.NoError(t, err)
	assert.Equal(t, "200.00", got.UnitPrice)
	assert.False(t, got.Active)
}
