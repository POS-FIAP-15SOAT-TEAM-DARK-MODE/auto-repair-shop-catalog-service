package interfaces

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
)

type SupplyHTTPController interface {
	Create(ctx context.Context, req *http.Request) (adapters.SupplyResponse, error)
	List(ctx context.Context, req *http.Request) (adapters.PaginatedSupplyResponse, error)
	Update(ctx context.Context, req *http.Request) (adapters.SupplyResponse, error)
	Delete(ctx context.Context, req *http.Request) error
}
