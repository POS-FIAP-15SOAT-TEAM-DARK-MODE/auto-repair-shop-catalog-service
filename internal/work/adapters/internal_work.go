package adapters

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"

// InternalWorkResponse is the payload of GET /internal/works/{id}, read by the Service Order service.
type InternalWorkResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UnitPrice   string `json:"unit_price"`
	Active      bool   `json:"active"`
}

func InternalWorkDomainToResponse(w domain.Work) InternalWorkResponse {
	return InternalWorkResponse{
		ID:          w.ID,
		Name:        w.Name,
		Description: w.Description,
		UnitPrice:   w.Price.StringFixed(2),
		Active:      w.Status.Bool(),
	}
}
