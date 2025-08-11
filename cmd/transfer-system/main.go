package main

import (
	"context"

	"github.com/tareqpi/transfer-system/internal/api"
	"github.com/tareqpi/transfer-system/internal/config"
	"github.com/tareqpi/transfer-system/internal/logger"
	"github.com/tareqpi/transfer-system/internal/repository"
	"github.com/tareqpi/transfer-system/internal/service"
	"go.uber.org/zap"
)

func main() {
	appConfig, err := config.Load()
	if err != nil {
		panic(err)
	}

	if err := logger.Init(appConfig.Environment); err != nil {
		panic(err)
	}
	defer logger.Sync()

	databasePool, err := repository.Setup(context.Background())
	if err != nil {
		logger.L().Fatal("repository setup failed", zap.Error(err))
	}

	postgresRepository := repository.NewPGRepository(databasePool)
	applicationService := service.NewService(postgresRepository)
	api.Setup(applicationService)
}
