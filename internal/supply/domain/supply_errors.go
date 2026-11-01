package domain

import "errors"

// Supply errors
var (
	ErrInvalidSupplyID            = errors.New("invalid supply id")
	ErrInvalidSupplyName          = errors.New("supply name must be at least 3 characters long")
	ErrInvalidSupplyDescription   = errors.New("supply description must be at least 10 characters long")
	ErrInvalidSupplyUnitPrice     = errors.New("supply unit price must be greater than 0")
	ErrInvalidSupplyStockQuantity = errors.New("supply stock quantity must be greater than or equal to 0")
	ErrInvalidSupplyVersion       = errors.New("supply version must be greater than or equal to 0")
	ErrInvalidSupplyAmount        = errors.New("supply amount must be greater than 0")
	ErrInvalidSupplyId            = errors.New("invalid supply id")
	ErrSupplyNotFound             = errors.New("supply not found")
	ErrSupplyOutOfStock           = errors.New("supply out of stock or requested quantity exceeds available stock")
)
