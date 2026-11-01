package adapters

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/shopspring/decimal"
)

type CreateSupply struct {
	Name          string          `json:"name" binding:"required"`
	Description   string          `json:"description" binding:"required"`
	UnitPrice     decimal.Decimal `json:"unitPrice" binding:"required"`
	StockQuantity int             `json:"stockQuantity" binding:"required"`
}

// UpdateSupply is the body of PUT /v1/supplies/:id. Version is the version the client
// read (GET returns it); the update applies only if the stored version still matches.
type UpdateSupply struct {
	Name          string          `json:"name" binding:"required"`
	Description   string          `json:"description" binding:"required"`
	UnitPrice     decimal.Decimal `json:"unitPrice" binding:"required"`
	StockQuantity int             `json:"stockQuantity" binding:"required"`
	Version       int             `json:"version" binding:"required"`
}

type SupplyResponse struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	StockQuantity int             `json:"stock_quantity"`
	Version       int             `json:"version"`
}

func SupplyDomainToResponse(supply domain.Supply) SupplyResponse {
	return SupplyResponse{
		ID:            supply.ID,
		Name:          supply.Name,
		Description:   supply.Description,
		UnitPrice:     supply.UnitPrice,
		StockQuantity: supply.StockQuantity,
		Version:       supply.Version,
	}
}
