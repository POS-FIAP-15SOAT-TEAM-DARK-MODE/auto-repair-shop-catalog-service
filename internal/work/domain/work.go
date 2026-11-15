package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type (
	Work struct {
		ID          string
		Name        string
		Description string
		Price       decimal.Decimal
		Status      WorkStatus
	}

	WorkStatus bool
)

const (
	ACTIVE   WorkStatus = true
	INACTIVE WorkStatus = false

	ActiveString   = "ACTIVE"
	InactiveString = "INACTIVE"

	nameMinSize        = 3
	descriptionMinSize = 10
)

func NewWork(name, description, rawPrice string, status WorkStatus) (*Work, error) {
	price, err := decimal.NewFromString(rawPrice)
	if err != nil {
		return nil, err
	}

	return &Work{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Price:       price,
		Status:      status,
	}, nil
}

func (s WorkStatus) String() string {
	if s {
		return ActiveString
	}
	return InactiveString
}

func (s WorkStatus) Bool() bool {
	return bool(s)
}

func BoolToWorkStatus(v bool) WorkStatus {
	if v {
		return ACTIVE
	}
	return INACTIVE
}

func StringToWorkStatus(v string) WorkStatus {
	if strings.ToUpper(v) == ActiveString {
		return ACTIVE
	}
	return INACTIVE
}

func ValidWorkStatusStringValue(v string) error {
	upper := strings.ToUpper(v)
	if upper != ActiveString && upper != InactiveString {
		return ErrInvalidWorkStatusValue
	}
	return nil
}

func (s *Work) Validate() error {
	var errs []error

	s.Name = strings.TrimSpace(s.Name)
	if s.Name == "" {
		errs = append(errs, ErrEmptyWorkName)
	} else if len(s.Name) < nameMinSize {
		errs = append(errs, ErrWorkNameShorterThenRequired)
	}

	s.Description = strings.TrimSpace(s.Description)
	if s.Description == "" {
		errs = append(errs, ErrEmptyWorkDescription)
	} else if len(s.Description) < descriptionMinSize {
		errs = append(errs, ErrWorkDescriptionShorterThenRequired)
	}

	if s.Price.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, ErrWorkPriceLessThenOrEqualZero)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
