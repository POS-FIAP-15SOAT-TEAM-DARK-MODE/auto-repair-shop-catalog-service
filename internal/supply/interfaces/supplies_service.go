package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=SupplyService --with-expecter
type SupplyService interface {
	Create(ctx context.Context, c adapters.CreateSupply) (adapters.SupplyResponse, error)
	Update(context.Context, string, adapters.UpdateSupply) (adapters.SupplyResponse, error)
	Delete(context.Context, string) error
	List(context.Context, adapters.ListSuppliesParams) (adapters.PaginatedSupplyResponse, error)
	FindById(context.Context, string) (domain.Supply, error)
	DecrementStockQuantity(context.Context, string, int) error
	IncrementStockQuantity(context.Context, string, int) error
}
