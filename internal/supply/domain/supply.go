package domain

import (
	"errors"

	"github.com/shopspring/decimal"
)

func NewSupply(id, name, description string, unitPrice decimal.Decimal, stockQuantity int) *Supply {
	return &Supply{
		ID:            id,
		Name:          name,
		Description:   description,
		UnitPrice:     unitPrice,
		StockQuantity: stockQuantity,
		Version:       1,
	}
}

type Supply struct {
	ID            string
	Name          string
	Description   string
	UnitPrice     decimal.Decimal
	StockQuantity int
	Version       int
}

func (s *Supply) IsValidName() error {
	if len(s.Name) < 3 {
		return ErrInvalidSupplyName
	}
	return nil
}

func (s *Supply) IsValidDescription() error {
	if len(s.Description) < 10 {
		return ErrInvalidSupplyDescription
	}
	return nil
}

func (s *Supply) IsValidUnitPrice() error {
	if s.UnitPrice.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidSupplyUnitPrice
	}
	return nil
}

func (s *Supply) IsValidStockQuantity() error {
	if s.StockQuantity < 0 {
		return ErrInvalidSupplyStockQuantity
	}
	return nil
}

func (s *Supply) IsValidVersion() error {
	if s.Version < 0 {
		return ErrInvalidSupplyVersion
	}
	return nil
}

func (s *Supply) Validate() error {
	var errs []error
	if err := s.IsValidName(); err != nil {
		errs = append(errs, err)
	}
	if err := s.IsValidDescription(); err != nil {
		errs = append(errs, err)
	}
	if err := s.IsValidUnitPrice(); err != nil {
		errs = append(errs, err)
	}
	if err := s.IsValidStockQuantity(); err != nil {
		errs = append(errs, err)
	}
	if err := s.IsValidVersion(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
