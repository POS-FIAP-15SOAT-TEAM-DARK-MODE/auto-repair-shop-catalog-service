package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=WorkService --with-expecter
type WorkService interface {
	Create(ctx context.Context, req adapters.CreateWork) (adapters.WorkResponse, error)
	List(ctx context.Context, params adapters.ListWorksParams) (adapters.PaginatedWorkResponse, error)
	Update(ctx context.Context, id string, req adapters.CreateWork) (adapters.WorkResponse, error)
	Delete(ctx context.Context, id string) error
}
