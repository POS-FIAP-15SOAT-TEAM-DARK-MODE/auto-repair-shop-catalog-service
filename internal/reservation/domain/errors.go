package domain

import "errors"

var (
	ErrInvalidServiceOrderID      = errors.New("invalid service order id")
	ErrInvalidReservationQuantity = errors.New("reservation quantity must be greater than zero")
	ErrReservationNotFound        = errors.New("reservation not found")
	ErrReservationAlreadyExists   = errors.New("supply already reserved for this service order")
	ErrInsufficientStock          = errors.New("insufficient stock for the requested quantity")
)
