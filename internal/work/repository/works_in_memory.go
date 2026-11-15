package repository

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/domain"
)

type inMemoryRepository struct {
	data map[string]domain.Work
}

func NewInMemory() *inMemoryRepository {
	return &inMemoryRepository{data: make(map[string]domain.Work)}
}

func (r *inMemoryRepository) Save(_ context.Context, w *domain.Work) error {
	r.data[w.ID] = *w
	return nil
}

func (r *inMemoryRepository) FindByID(_ context.Context, id string) (domain.Work, error) {
	w, ok := r.data[id]
	if !ok {
		return domain.Work{}, domain.ErrWorkNotFound
	}
	return w, nil
}

func (r *inMemoryRepository) Count(_ context.Context, params adapters.ListWorksParams) (int64, error) {
	if params.Status == "" {
		return int64(len(r.data)), nil
	}

	status := domain.StringToWorkStatus(params.Status)
	var total int64
	for _, v := range r.data {
		if v.Status == status {
			total++
		}
	}
	return total, nil
}

func (r *inMemoryRepository) Search(_ context.Context, params adapters.ListWorksParams) ([]domain.Work, error) {
	var filtered []domain.Work
	for _, v := range r.data {
		if params.Status != "" && v.Status != domain.StringToWorkStatus(params.Status) {
			continue
		}
		filtered = append(filtered, v)
	}

	offset := int((params.Page - 1) * params.PageSize)
	if offset >= len(filtered) {
		return []domain.Work{}, nil
	}

	result := make([]domain.Work, 0, params.PageSize)
	for i := offset; i < len(filtered) && int64(len(result)) < params.PageSize; i++ {
		result = append(result, filtered[i])
	}
	return result, nil
}

func (r *inMemoryRepository) Delete(_ context.Context, id string) error {
	delete(r.data, id)
	return nil
}
