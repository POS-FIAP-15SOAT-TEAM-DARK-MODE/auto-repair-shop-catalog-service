package adapters

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
)

type ListWorksParams struct {
	PageSize int64
	Page     int64
	Status   string
}

type PaginatedWorkResponse app.PaginatedResponse[WorkResponse]

func WorksDomainToResponse(works []domain.Work) []WorkResponse {
	res := make([]WorkResponse, len(works))
	for i, w := range works {
		res[i] = WorkDomainToResponse(w)
	}
	return res
}
