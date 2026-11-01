package repository

import (
	"context"
	"fmt"
	"slices"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
)

type inMemory struct {
	data map[string]domain.Supply
}

func NewInMemory() *inMemory {
	return &inMemory{
		make(map[string]domain.Supply),
	}
}

func (i *inMemory) Save(_ context.Context, supply *domain.Supply) error {
	i.data[supply.ID] = *supply
	return nil
}

func (i *inMemory) Delete(_ context.Context, id string) error {
	delete(i.data, id)
	return nil
}

func (i *inMemory) Search(_ context.Context, params adapters.ListSuppliesParams) ([]domain.Supply, error) {
	var filtered []domain.Supply
	for _, v := range i.data {
		if params.Version != "" && fmt.Sprintf("%v", v.Version) != params.Version {
			continue
		}

		if params.ID != "" && v.ID != params.ID {
			continue
		}
		filtered = append(filtered, v)
	}
	slices.SortFunc(filtered, func(a, b domain.Supply) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	start := int((params.Page - 1) * params.PageSize)
	if start >= len(filtered) {
		return []domain.Supply{}, nil
	}
	end := start + int(params.PageSize)
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], nil
}

func (i *inMemory) Count(_ context.Context, params adapters.ListSuppliesParams) (int64, error) {
	var total int64
	for _, v := range i.data {
		if params.Version != "" && fmt.Sprintf("%v", v.Version) != params.Version {
			continue
		}

		if params.ID != "" && v.ID != params.ID {
			continue
		}

		total++
	}
	return total, nil
}
