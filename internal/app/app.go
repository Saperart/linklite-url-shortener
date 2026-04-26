package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Saperart/linklite-url-shortener/internal/config"
	httpcontroller "github.com/Saperart/linklite-url-shortener/internal/controller/http"
	"github.com/Saperart/linklite-url-shortener/internal/entity"
	"github.com/Saperart/linklite-url-shortener/internal/repository/memory"
	"github.com/Saperart/linklite-url-shortener/internal/repository/postgres"
	gen "github.com/Saperart/linklite-url-shortener/internal/usecase/generator"
	"github.com/Saperart/linklite-url-shortener/internal/usecase/link_service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	cfg       config.Config
	logger    *zap.Logger
	server    *http.Server
	pool      *pgxpool.Pool
	closeOnce sync.Once
}

func New(cfg config.Config, logger *zap.Logger) (*App, error) {
	repo, pool, err := buildRepository(cfg)
	if err != nil {
		return nil, fmt.Errorf("build repository: %w", err)
	}

	generator, err := gen.NewRandomCodeGenerator(
		entity.ShortCodeLength,
		entity.ShortCodeAlphabet,
	)
	if err != nil {
		if pool != nil {
			pool.Close()
		}
		return nil, fmt.Errorf("create random code generator: %w", err)
	}

	service := link_service.NewLinkService(
		repo,
		generator,
	)

	handler := httpcontroller.NewHandler(
		service,
		logger,
		cfg.BaseURL,
	)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler.Router(),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	return &App{
		cfg:    cfg,
		logger: logger,
		server: server,
		pool:   pool,
	}, nil
}

func buildRepository(cfg config.Config) (link_service.LinkRepository, *pgxpool.Pool, error) {
	switch cfg.StorageType {
	case config.StorageTypeMemory:
		return memory.NewRepository(), nil, nil
	case config.StorageTypePostgres:
		ctx, cancel := context.WithTimeout(context.Background(), cfg.PostgresConnectTimeout)
		defer cancel()

		pool, err := pgxpool.New(ctx, cfg.DatabaseDSN)
		if err != nil {
			return nil, nil, fmt.Errorf("create pgx pool: %w", err)
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return nil, nil, fmt.Errorf("ping postgres: %w", err)
		}

		return postgres.NewRepository(pool), pool, nil

	default:
		return nil, nil, fmt.Errorf("unknown storage type: %s", cfg.StorageType)
	}
}
