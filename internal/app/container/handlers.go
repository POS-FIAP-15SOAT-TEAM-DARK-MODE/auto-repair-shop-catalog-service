package container

import (
	reservationInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/interfaces"
	supplyInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply/interfaces"
	workInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work/interfaces"

	"github.com/gin-gonic/gin"
)

type HandlersWrapper struct {
	WorkHandler         workInterfaces.WorkHTTPController
	WorkInternalHandler workInterfaces.WorkInternalHTTPController
	SupplyHandler       supplyInterfaces.SupplyHTTPController
	ReservationHandler  reservationInterfaces.ReservationHTTPController
}

type Middlewares map[string]gin.HandlerFunc
