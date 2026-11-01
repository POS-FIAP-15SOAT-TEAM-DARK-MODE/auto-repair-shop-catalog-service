package bootstrap

import (
	"context"
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/middleware"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/routing"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/messaging"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/reservation/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/supply"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/work"
	"github.com/gin-gonic/gin"
)

// HTTPServer builds the HTTP server (public /v1 API and /internal API).
func HTTPServer() app.Server {
	middlewares := middlewaresContainer()
	db := app.DBConnect()
	app.RunSeed(context.Background(), db)
	router := NewRouter(db, env.GetString("INTERNAL_TOKEN", ""), middlewares)

	port := env.GetString("PORT", "8082")
	return app.NewHTTPServer("http-api", ":"+port, router)
}

// NewRouter wires the handlers of every catalog domain onto a gin engine.
func NewRouter(db *sql.DB, internalToken string, middlewares *container.Middlewares) *gin.Engine {
	return routing.SetupRouter(httpContainer(db), middlewares, internalToken)
}

// Consumer returns the SQS consumer for ReleaseStock commands, or nil when
// CATALOG_QUEUE_URL is empty (the service then runs without messaging).
func Consumer(ctx context.Context) (app.Server, error) {
	queueURL := env.GetString("CATALOG_QUEUE_URL", "")
	if queueURL == "" {
		return nil, nil
	}

	client, err := messaging.NewSQSClient(ctx)
	if err != nil {
		return nil, err
	}

	db := app.DBConnect()
	unitOfWork := uow.NewTransactionalUoW(db)
	svc := reservation.NewService(unitOfWork, repository.NewPostgresReservations(), repository.NewPostgresStock())
	handlers := map[string]messaging.Handler{
		"ReleaseStock": reservation.ReleaseStockHandler(svc),
	}
	return messaging.NewConsumer(queueURL, client, unitOfWork, messaging.NewPostgresDedupe(), handlers), nil
}

func middlewaresContainer() *container.Middlewares {
	return &container.Middlewares{
		"Logger":       middleware.Logger(),
		"Recovery":     middleware.Recovery(),
		"Metrics":      middleware.Metrics(),
		"ErrorHandler": middleware.ErrorHandler(),
	}
}

func httpContainer(db *sql.DB) *container.HandlersWrapper {
	return &container.HandlersWrapper{
		WorkHandler:         work.NewHTTPController(db),
		WorkInternalHandler: work.NewInternalHTTPController(db),
		SupplyHandler:       supply.NewHTTPController(db),
		ReservationHandler:  reservation.NewHTTPController(db),
	}
}
