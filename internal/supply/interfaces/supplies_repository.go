package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=SupplyRepository --with-expecter
type SupplyRepository interface {
	Save(context.Context, *domain.Supply) error
	Delete(context.Context, string) error
	Search(context.Context, adapters.ListSuppliesParams) ([]domain.Supply, error)
	Count(context.Context, adapters.ListSuppliesParams) (int64, error)
}
