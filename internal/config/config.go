package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	StorageTypeMemory   = "memory"
	StorageTypePostgres = "postgres"

	AppEnvLocal = "local"
	AppEnvDev   = "dev"
	AppEnvProd  = "prod"
)

type Config struct {
	AppEnv                 string        `env:"APP_ENV" envDefault:"local"`
	StorageType            string        `env:"STORAGE_TYPE" envDefault:"memory"`
	HTTPAddr               string        `env:"HTTP_ADDR"`
	BaseURL                string        `env:"BASE_URL"`
	DatabaseDSN            string        `env:"DATABASE_DSN"`
	ShutdownTimeout        time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	ReadHeaderTimeout      time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"`
	ReadTimeout            time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout           time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout            time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	PostgresConnectTimeout time.Duration `env:"POSTGRES_CONNECT_TIMEOUT" envDefault:"5s"`
}

func Load() (Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse env: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.AppEnv != AppEnvLocal && c.AppEnv != AppEnvDev && c.AppEnv != AppEnvProd {
		return fmt.Errorf("invalid APP_ENV: %q", c.AppEnv)
	}

	if c.StorageType != StorageTypeMemory && c.StorageType != StorageTypePostgres {
		return fmt.Errorf("invalid STORAGE_TYPE: %q", c.StorageType)
	}

	if strings.TrimSpace(c.HTTPAddr) == "" {
		return errors.New("HTTP_ADDR must not be empty")
	}

	if strings.TrimSpace(c.BaseURL) == "" {
		return errors.New("BASE_URL must not be empty")
	}

	if c.StorageType == StorageTypePostgres && strings.TrimSpace(c.DatabaseDSN) == "" {
		return errors.New("DATABASE_DSN is required for postgres storage")
	}

	if c.ShutdownTimeout <= 0 {
		return errors.New("SHUTDOWN_TIMEOUT must be positive")
	}

	if c.ReadHeaderTimeout <= 0 {
		return errors.New("HTTP_READ_HEADER_TIMEOUT must be positive")
	}

	if c.ReadTimeout <= 0 {
		return errors.New("HTTP_READ_TIMEOUT must be positive")
	}

	if c.WriteTimeout <= 0 {
		return errors.New("HTTP_WRITE_TIMEOUT must be positive")
	}

	if c.IdleTimeout <= 0 {
		return errors.New("HTTP_IDLE_TIMEOUT must be positive")
	}

	if c.PostgresConnectTimeout <= 0 {
		return errors.New("POSTGRES_CONNECT_TIMEOUT must be positive")
	}

	return nil
}
