package interfaces

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=WorkInternalService --with-expecter
type WorkInternalService interface {
	FindByID(ctx context.Context, id string) (adapters.InternalWorkResponse, error)
}

type WorkInternalHTTPController interface {
	Get(ctx context.Context, req *http.Request) (adapters.InternalWorkResponse, error)
}
