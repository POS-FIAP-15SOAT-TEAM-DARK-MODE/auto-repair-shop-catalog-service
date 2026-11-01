package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=WorkRepository --with-expecter
type WorkRepository interface {
	Save(ctx context.Context, work *domain.Work) error
	FindByID(ctx context.Context, id string) (domain.Work, error)
	Search(ctx context.Context, params adapters.ListWorksParams) ([]domain.Work, error)
	Count(ctx context.Context, params adapters.ListWorksParams) (int64, error)
	Delete(ctx context.Context, id string) error
}
