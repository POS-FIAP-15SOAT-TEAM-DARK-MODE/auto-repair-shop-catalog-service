package adapters

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"

type CreateWork struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Status      string `json:"status"`
}

type WorkResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Status      string `json:"status"`
}

func WorkDomainToResponse(work domain.Work) WorkResponse {
	return WorkResponse{
		ID:          work.ID,
		Name:        work.Name,
		Description: work.Description,
		Price:       work.Price.String(),
		Status:      work.Status.String(),
	}
}
