package main

import (
	"context"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/app/bootstrap"
	pkgApp "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/app"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	defer logger.Global().Sync()

	manager := pkgApp.NewLifecycleManager(5 * time.Second)
	manager.Add(bootstrap.HTTPServer())

	consumer, err := bootstrap.Consumer(context.Background())
	if err != nil {
		logger.Global().Fatal("failed to create SQS consumer", zap.Error(err))
	}
	if consumer != nil {
		manager.Add(consumer)
	}

	if err := manager.Run(context.Background()); err != nil {
		logger.Global().Fatal("application failed")
	}
}
