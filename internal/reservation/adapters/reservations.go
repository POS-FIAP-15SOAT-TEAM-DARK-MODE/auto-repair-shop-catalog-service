package adapters

// ReserveRequest is the body of POST /internal/reservations.
type ReserveRequest struct {
	ServiceOrderID string `json:"service_order_id"`
	SupplyID       string `json:"supply_id"`
	Quantity       int    `json:"quantity"`
}

// ReserveResponse is returned with 201 and carries the snapshot the caller stores.
type ReserveResponse struct {
	ReservationID string `json:"reservation_id"`
	SupplyID      string `json:"supply_id"`
	Name          string `json:"name"`
	UnitPrice     string `json:"unit_price"`
}
