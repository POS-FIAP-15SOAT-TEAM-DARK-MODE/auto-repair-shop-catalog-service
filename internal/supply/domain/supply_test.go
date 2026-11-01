package domain_test

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// ─── NewSupply ────────────────────────────────────────────────────────────────

func TestNewSupply(t *testing.T) {
	name := "Brake Pad"
	description := "High performance brake pad"
	unitPrice := decimal.NewFromFloat(49.99)
	stockQuantity := 10
	version := 1
	id := "s-1"

	s := domain.NewSupply(id, name, description, unitPrice, stockQuantity)

	assert.Equal(t, id, s.ID)
	assert.Equal(t, name, s.Name)
	assert.Equal(t, description, s.Description)
	assert.True(t, unitPrice.Equal(s.UnitPrice))
	assert.Equal(t, stockQuantity, s.StockQuantity)
	assert.Equal(t, version, s.Version)
}

// ─── IsValidName ──────────────────────────────────────────────────────────────

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid name", "Brake Pad", nil},
		{"exactly 3 chars", "Pad", nil},
		{"2 chars is invalid", "Pa", domain.ErrInvalidSupplyName},
		{"empty string is invalid", "", domain.ErrInvalidSupplyName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &domain.Supply{Name: tt.input}
			assert.ErrorIs(t, s.IsValidName(), tt.wantErr)
		})
	}
}

// ─── IsValidDescription ───────────────────────────────────────────────────────

func TestIsValidDescription(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid description", "High performance brake pad", nil},
		{"exactly 10 chars", "1234567890", nil},
		{"9 chars is invalid", "123456789", domain.ErrInvalidSupplyDescription},
		{"empty string is invalid", "", domain.ErrInvalidSupplyDescription},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &domain.Supply{Description: tt.input}
			assert.ErrorIs(t, s.IsValidDescription(), tt.wantErr)
		})
	}
}

// ─── IsValidUnitPrice ─────────────────────────────────────────────────────────

func TestIsValidUnitPrice(t *testing.T) {
	tests := []struct {
		name    string
		input   decimal.Decimal
		wantErr error
	}{
		{"positive price", decimal.NewFromFloat(49.99), nil},
		{"minimum positive price", decimal.NewFromFloat(0.01), nil},
		{"zero is invalid", decimal.Zero, domain.ErrInvalidSupplyUnitPrice},
		{"negative is invalid", decimal.NewFromFloat(-1), domain.ErrInvalidSupplyUnitPrice},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &domain.Supply{UnitPrice: tt.input}
			assert.ErrorIs(t, s.IsValidUnitPrice(), tt.wantErr)
		})
	}
}

// ─── IsValidStockQuantity ─────────────────────────────────────────────────────

func TestIsValidStockQuantity(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr error
	}{
		{"positive quantity", 10, nil},
		{"zero is valid (out of stock)", 0, nil},
		{"negative is invalid", -1, domain.ErrInvalidSupplyStockQuantity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &domain.Supply{StockQuantity: tt.input}
			assert.ErrorIs(t, s.IsValidStockQuantity(), tt.wantErr)
		})
	}
}

// ─── IsValidVersion ───────────────────────────────────────────────────────────

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr error
	}{
		{"positive version", 1, nil},
		{"zero is valid", 0, nil},
		{"negative is invalid", -1, domain.ErrInvalidSupplyVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &domain.Supply{Version: tt.input}
			assert.ErrorIs(t, s.IsValidVersion(), tt.wantErr)
		})
	}
}

// ─── Validate ─────────────────────────────────────────────────────────────────

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		supply  *domain.Supply
		wantErr []error // all errors expected to be present in the joined error
	}{
		{
			name: "valid supply",
			supply: &domain.Supply{
				Name:          "Brake Pad",
				Description:   "High performance brake pad",
				UnitPrice:     decimal.NewFromFloat(49.99),
				StockQuantity: 10,
				Version:       1,
			},
		},
		{
			name:    "invalid name",
			supply:  &domain.Supply{Name: "Pa", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 1},
			wantErr: []error{domain.ErrInvalidSupplyName},
		},
		{
			name:    "invalid description",
			supply:  &domain.Supply{Name: "Brake Pad", Description: "short", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: 1},
			wantErr: []error{domain.ErrInvalidSupplyDescription},
		},
		{
			name:    "invalid unit price",
			supply:  &domain.Supply{Name: "Brake Pad", Description: "High performance brake pad", UnitPrice: decimal.Zero, StockQuantity: 10, Version: 1},
			wantErr: []error{domain.ErrInvalidSupplyUnitPrice},
		},
		{
			name:    "invalid stock quantity",
			supply:  &domain.Supply{Name: "Brake Pad", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: -1, Version: 1},
			wantErr: []error{domain.ErrInvalidSupplyStockQuantity},
		},
		{
			name:    "invalid version",
			supply:  &domain.Supply{Name: "Brake Pad", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10, Version: -1},
			wantErr: []error{domain.ErrInvalidSupplyVersion},
		},
		{
			name:    "multiple errors are joined",
			supply:  &domain.Supply{},
			wantErr: []error{domain.ErrInvalidSupplyName, domain.ErrInvalidSupplyDescription, domain.ErrInvalidSupplyUnitPrice},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.supply.Validate()

			if len(tt.wantErr) == 0 {
				assert.NoError(t, err)
				return
			}

			for _, expected := range tt.wantErr {
				assert.ErrorIs(t, err, expected)
			}
		})
	}
}
