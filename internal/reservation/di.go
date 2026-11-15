package reservation

import (
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/repository"
)

func NewHTTPController(db *sql.DB) interfaces.ReservationHTTPController {
	unitOfWork := uow.NewTransactionalUoW(db)
	svc := NewService(unitOfWork, repository.NewPostgresReservations(), repository.NewPostgresStock())
	return NewController(svc)
}
