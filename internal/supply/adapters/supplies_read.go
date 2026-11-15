package adapters

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
)

type ListSuppliesParams struct {
	PageSize int64
	Page     int64
	Version  string
	ID       string
}

type PaginatedSupplyResponse app.PaginatedResponse[SupplyResponse]

func SuppliesDomainToResponse(supplies []domain.Supply) []SupplyResponse {
	res := make([]SupplyResponse, len(supplies))
	for i, supply := range supplies {
		res[i] = SupplyDomainToResponse(supply)
	}
	return res
}
