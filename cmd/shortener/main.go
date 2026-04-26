package main

import (
	"log"

	"github.com/Saperart/linklite-url-shortener/internal/app"
	"github.com/Saperart/linklite-url-shortener/internal/config"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("can not initialize config: %s", err)
	}

	logger, err := newLogger(cfg.AppEnv)
	if err != nil {
		log.Fatalf("can not initialize logger: %s", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal("can not initialize app", zap.Error(err))
	}

	if err := application.Run(); err != nil {
		logger.Fatal("application failed", zap.Error(err))
	}
}

func newLogger(appEnv string) (*zap.Logger, error) {
	if appEnv == "local" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
