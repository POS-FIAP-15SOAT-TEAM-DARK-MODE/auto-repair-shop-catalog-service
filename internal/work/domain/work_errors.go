package domain

import "errors"

var (
	ErrEmptyWorkName                      = errors.New("work name can't be empty")
	ErrWorkNameShorterThenRequired        = errors.New("work name should have at least 3 letters")
	ErrEmptyWorkDescription               = errors.New("work description can't be empty")
	ErrWorkDescriptionShorterThenRequired = errors.New("work description should have at least 10 characters")
	ErrWorkPriceLessThenOrEqualZero       = errors.New("work price should be bigger then 0")
	ErrInvalidWorkPriceValue              = errors.New("invalid work price value")
	ErrInvalidWorkStatusValue             = errors.New("work status should be ACTIVE or INACTIVE")
	ErrInvalidWorkId                      = errors.New("invalid work id")
	ErrWorkNotFound                       = errors.New("work not found")
)
