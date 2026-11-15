package repository

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedInMemory(t *testing.T, repo *inMemory, supplies []*domain.Supply) {
	t.Helper()
	for _, v := range supplies {
		require.NoError(t, repo.Save(context.Background(), v))
	}
}

func TestInMemory_Save(t *testing.T) {
	tests := []struct {
		name            string
		supply          *domain.Supply
		update          func(*domain.Supply)
		wantName        string
		wantDescription string
	}{
		{
			name: "persists supply",
			supply: &domain.Supply{
				ID:            "supply-1",
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
			wantName:        "Brake Pad",
			wantDescription: "High performance brake pad",
		},
		{
			name: "overwrites existing supply",
			supply: &domain.Supply{
				ID:            "supply-1",
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
			update: func(v *domain.Supply) {
				v.Name = "fiat"
				v.Description = "uno"
			},
			wantName:        "fiat",
			wantDescription: "uno",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemory()
			ctx := context.Background()

			require.NoError(t, repo.Save(ctx, tt.supply))
			if tt.update != nil {
				tt.update(tt.supply)
				require.NoError(t, repo.Save(ctx, tt.supply))
			}

			got, err := repo.Search(ctx, adapters.ListSuppliesParams{
				ID:       tt.supply.ID,
				Page:     1,
				PageSize: 10,
			})
			require.NoError(t, err)
			require.Len(t, got, 1)
			assert.Equal(t, tt.wantName, got[0].Name)
			assert.Equal(t, tt.wantDescription, got[0].Description)
		})
	}
}

func TestInMemory_Delete(t *testing.T) {
	tests := []struct {
		name      string
		seed      *domain.Supply
		deleteID  string
		wantCount int64
	}{
		{
			name: "deletes existing vehicle",
			seed: &domain.Supply{
				ID:            "supply-1",
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
			deleteID:  "supply-1",
			wantCount: 0,
		},
		{
			name:      "non-existent id does not error",
			deleteID:  "missing-id",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemory()
			ctx := context.Background()

			if tt.seed != nil {
				require.NoError(t, repo.Save(ctx, tt.seed))
			}

			err := repo.Delete(ctx, tt.deleteID)
			require.NoError(t, err)

			got, err := repo.Search(ctx, adapters.ListSuppliesParams{
				ID:       tt.deleteID,
				Page:     1,
				PageSize: 10,
			})
			require.NoError(t, err)
			assert.Empty(t, got)

			total, err := repo.Count(ctx, adapters.ListSuppliesParams{})
			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, total)
		})
	}
}

func TestInMemory_Search(t *testing.T) {
	filterSeed := []*domain.Supply{
		{ID: "supply-1", Name: "Brake Pad", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 1},
		{ID: "supply-2", Name: "Pad Brake", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 2},
		{ID: "supply-3", Name: "Brake Pedal", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 3},
	}

	paginationSeed := []*domain.Supply{
		{ID: "supply-1", Name: "Brake Pad", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 1},
		{ID: "supply-2", Name: "Pad Brake", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 1},
		{ID: "supply-3", Name: "Brake Pedal", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 1},
	}

	tests := []struct {
		name   string
		seed   []*domain.Supply
		params adapters.ListSuppliesParams
		want   []string
	}{
		{
			name: "filter by version",
			seed: filterSeed,
			params: adapters.ListSuppliesParams{
				Version:  "2",
				Page:     1,
				PageSize: 10,
			},
			want: []string{"supply-2"},
		},
		{
			name: "filter by supply id",
			seed: filterSeed,
			params: adapters.ListSuppliesParams{
				ID:       "supply-3",
				Page:     1,
				PageSize: 10,
			},
			want: []string{"supply-3"},
		},
		{
			name: "combined filters",
			seed: filterSeed,
			params: adapters.ListSuppliesParams{
				ID:       "supply-1",
				Version:  "1",
				Page:     1,
				PageSize: 10,
			},
			want: []string{"supply-1"},
		},
		{
			name: "combined filters no matches",
			seed: filterSeed,
			params: adapters.ListSuppliesParams{
				ID:       "supply-1",
				Version:  "2",
				Page:     1,
				PageSize: 10,
			},
			want: []string{},
		},
		{
			name: "no matches",
			seed: filterSeed,
			params: adapters.ListSuppliesParams{
				Version:  "NOTFOUND",
				Page:     1,
				PageSize: 10,
			},
			want: []string{},
		},
		{
			name: "first page",
			seed: paginationSeed,
			params: adapters.ListSuppliesParams{
				Version:  "1",
				Page:     1,
				PageSize: 2,
			},
			want: []string{"supply-1", "supply-2"},
		},
		{
			name: "second page",
			seed: paginationSeed,
			params: adapters.ListSuppliesParams{
				Version:  "1",
				Page:     2,
				PageSize: 2,
			},
			want: []string{"supply-3"},
		},
		{
			name: "page beyond range",
			seed: paginationSeed,
			params: adapters.ListSuppliesParams{
				ID:       "supply-1",
				Page:     3,
				PageSize: 2,
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemory()
			seedInMemory(t, repo, tt.seed)

			got, err := repo.Search(context.Background(), tt.params)
			require.NoError(t, err)

			ids := make([]string, len(got))
			for i, v := range got {
				ids[i] = v.ID
			}
			assert.Equal(t, tt.want, ids)
		})
	}
}

func TestInMemory_Count(t *testing.T) {
	seed := []*domain.Supply{
		{ID: "supply-1", Name: "Brake Pad", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 1},
		{ID: "supply-2", Name: "Pad Brake", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 1},
		{ID: "supply-3", Name: "Brake Pedal", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 2},
	}

	tests := []struct {
		name   string
		params adapters.ListSuppliesParams
		want   int64
	}{
		{
			name:   "counts all vehicles without filters",
			params: adapters.ListSuppliesParams{},
			want:   3,
		},
		{
			name: "counts by version id",
			params: adapters.ListSuppliesParams{
				Version: "1",
			},
			want: 2,
		},
		{
			name: "counts by ID",
			params: adapters.ListSuppliesParams{
				ID: "supply-1",
			},
			want: 1,
		},
		{
			name: "returns zero when no match",
			params: adapters.ListSuppliesParams{
				Version: "NOTFOUND",
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewInMemory()
			seedInMemory(t, repo, seed)

			total, err := repo.Count(context.Background(), tt.params)
			require.NoError(t, err)
			assert.Equal(t, tt.want, total)
		})
	}
}
