package interfaces

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
)

type WorkHTTPController interface {
	Create(ctx context.Context, req *http.Request) (adapters.WorkResponse, error)
	List(ctx context.Context, req *http.Request) (adapters.PaginatedWorkResponse, error)
	Update(ctx context.Context, req *http.Request) (adapters.WorkResponse, error)
	Delete(ctx context.Context, req *http.Request) error
}
