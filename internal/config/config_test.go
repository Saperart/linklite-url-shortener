package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testHTTPAddr    = ":8080"
	testBaseURL     = "http://localhost:8080"
	testDatabaseDSN = "postgres://shortener:shortener@localhost:5432/shortener?sslmode=disable"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		prepare   func(t *testing.T)
		wantCfg   Config
		wantErr   bool
		errSubstr string
	}{
		{
			name: "memory storage config",
			prepare: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("STORAGE_TYPE", StorageTypeMemory)
				t.Setenv("DATABASE_DSN", "")
			},
			wantCfg: Config{
				AppEnv:                 AppEnvLocal,
				StorageType:            StorageTypeMemory,
				HTTPAddr:               testHTTPAddr,
				BaseURL:                testBaseURL,
				DatabaseDSN:            "",
				ShutdownTimeout:        10 * time.Second,
				ReadHeaderTimeout:      5 * time.Second,
				ReadTimeout:            10 * time.Second,
				WriteTimeout:           10 * time.Second,
				IdleTimeout:            60 * time.Second,
				PostgresConnectTimeout: 5 * time.Second,
			},
		},
		{
			name: "postgres storage config",
			prepare: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("STORAGE_TYPE", StorageTypePostgres)
				t.Setenv("DATABASE_DSN", testDatabaseDSN)
			},
			wantCfg: Config{
				AppEnv:                 AppEnvLocal,
				StorageType:            StorageTypePostgres,
				HTTPAddr:               testHTTPAddr,
				BaseURL:                testBaseURL,
				DatabaseDSN:            testDatabaseDSN,
				ShutdownTimeout:        10 * time.Second,
				ReadHeaderTimeout:      5 * time.Second,
				ReadTimeout:            10 * time.Second,
				WriteTimeout:           10 * time.Second,
				IdleTimeout:            60 * time.Second,
				PostgresConnectTimeout: 5 * time.Second,
			},
		},
		{
			name: "custom timeouts",
			prepare: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("SHUTDOWN_TIMEOUT", "20s")
				t.Setenv("HTTP_READ_HEADER_TIMEOUT", "2s")
				t.Setenv("HTTP_READ_TIMEOUT", "3s")
				t.Setenv("HTTP_WRITE_TIMEOUT", "4s")
				t.Setenv("HTTP_IDLE_TIMEOUT", "30s")
				t.Setenv("POSTGRES_CONNECT_TIMEOUT", "7s")
			},
			wantCfg: Config{
				AppEnv:                 AppEnvLocal,
				StorageType:            StorageTypeMemory,
				HTTPAddr:               testHTTPAddr,
				BaseURL:                testBaseURL,
				DatabaseDSN:            "",
				ShutdownTimeout:        20 * time.Second,
				ReadHeaderTimeout:      2 * time.Second,
				ReadTimeout:            3 * time.Second,
				WriteTimeout:           4 * time.Second,
				IdleTimeout:            30 * time.Second,
				PostgresConnectTimeout: 7 * time.Second,
			},
		},
		{
			name: "invalid duration",
			prepare: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("SHUTDOWN_TIMEOUT", "bad-duration")
			},
			wantErr:   true,
			errSubstr: "parse env",
		},
		{
			name: "missing http addr",
			prepare: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("HTTP_ADDR", "")
			},
			wantErr:   true,
			errSubstr: "HTTP_ADDR must not be empty",
		},
		{
			name: "missing base url",
			prepare: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("BASE_URL", "")
			},
			wantErr:   true,
			errSubstr: "BASE_URL must not be empty",
		},
		{
			name: "postgres without dsn",
			prepare: func(t *testing.T) {
				setValidEnv(t)
				t.Setenv("STORAGE_TYPE", StorageTypePostgres)
				t.Setenv("DATABASE_DSN", "")
			},
			wantErr:   true,
			errSubstr: "DATABASE_DSN is required for postgres storage",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			tc.prepare(t)

			cfg, err := Load()

			if tc.wantErr {
				require.Error(t, err)
				if tc.errSubstr != "" {
					assert.Contains(t, err.Error(), tc.errSubstr)
				}
				return
			}

			require.NoError(t, err)

			assertConfigEqual(t, tc.wantCfg, cfg)
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cfg       Config
		wantErr   bool
		errSubstr string
	}{
		{
			name: "valid memory config",
			cfg:  validConfig(StorageTypeMemory),
		},
		{
			name: "valid postgres config",
			cfg:  validConfig(StorageTypePostgres),
		},
		{
			name: "invalid app env",
			cfg: func() Config {
				cfg := validConfig(StorageTypeMemory)
				cfg.AppEnv = "invalid"
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "invalid APP_ENV",
		},
		{
			name: "invalid storage type",
			cfg: func() Config {
				cfg := validConfig(StorageTypeMemory)
				cfg.StorageType = "redis"
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "invalid STORAGE_TYPE",
		},
		{
			name: "empty http addr",
			cfg: func() Config {
				cfg := validConfig(StorageTypeMemory)
				cfg.HTTPAddr = "   "
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "HTTP_ADDR must not be empty",
		},
		{
			name: "empty base url",
			cfg: func() Config {
				cfg := validConfig(StorageTypeMemory)
				cfg.BaseURL = "   "
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "BASE_URL must not be empty",
		},
		{
			name: "postgres without database dsn",
			cfg: func() Config {
				cfg := validConfig(StorageTypePostgres)
				cfg.DatabaseDSN = ""
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "DATABASE_DSN is required for postgres storage",
		},
		{
			name: "zero shutdown timeout",
			cfg: func() Config {
				cfg := validConfig(StorageTypeMemory)
				cfg.ShutdownTimeout = 0
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "SHUTDOWN_TIMEOUT must be positive",
		},
		{
			name: "negative read header timeout",
			cfg: func() Config {
				cfg := validConfig(StorageTypeMemory)
				cfg.ReadHeaderTimeout = -time.Second
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "HTTP_READ_HEADER_TIMEOUT must be positive",
		},
		{
			name: "zero read timeout",
			cfg: func() Config {
				cfg := validConfig(StorageTypeMemory)
				cfg.ReadTimeout = 0
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "HTTP_READ_TIMEOUT must be positive",
		},
		{
			name: "zero write timeout",
			cfg: func() Config {
				cfg := validConfig(StorageTypeMemory)
				cfg.WriteTimeout = 0
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "HTTP_WRITE_TIMEOUT must be positive",
		},
		{
			name: "zero idle timeout",
			cfg: func() Config {
				cfg := validConfig(StorageTypeMemory)
				cfg.IdleTimeout = 0
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "HTTP_IDLE_TIMEOUT must be positive",
		},
		{
			name: "zero postgres connect timeout",
			cfg: func() Config {
				cfg := validConfig(StorageTypeMemory)
				cfg.PostgresConnectTimeout = 0
				return cfg
			}(),
			wantErr:   true,
			errSubstr: "POSTGRES_CONNECT_TIMEOUT must be positive",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.cfg.validate()

			if tc.wantErr {
				require.Error(t, err)
				if tc.errSubstr != "" {
					assert.Contains(t, err.Error(), tc.errSubstr)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func setValidEnv(t *testing.T) {
	t.Helper()

	t.Setenv("APP_ENV", AppEnvLocal)
	t.Setenv("STORAGE_TYPE", StorageTypeMemory)
	t.Setenv("HTTP_ADDR", testHTTPAddr)
	t.Setenv("BASE_URL", testBaseURL)
	t.Setenv("DATABASE_DSN", "")

	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "5s")
	t.Setenv("HTTP_READ_TIMEOUT", "10s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "10s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "60s")
	t.Setenv("POSTGRES_CONNECT_TIMEOUT", "5s")
}

func validConfig(storageType string) Config {
	cfg := Config{
		AppEnv:                 AppEnvLocal,
		StorageType:            storageType,
		HTTPAddr:               testHTTPAddr,
		BaseURL:                testBaseURL,
		DatabaseDSN:            "",
		ShutdownTimeout:        10 * time.Second,
		ReadHeaderTimeout:      5 * time.Second,
		ReadTimeout:            10 * time.Second,
		WriteTimeout:           10 * time.Second,
		IdleTimeout:            60 * time.Second,
		PostgresConnectTimeout: 5 * time.Second,
	}

	if storageType == StorageTypePostgres {
		cfg.DatabaseDSN = testDatabaseDSN
	}

	return cfg
}

func assertConfigEqual(t *testing.T, want Config, got Config) {
	t.Helper()

	assert.Equal(t, want.AppEnv, got.AppEnv)
	assert.Equal(t, want.StorageType, got.StorageType)
	assert.Equal(t, want.HTTPAddr, got.HTTPAddr)
	assert.Equal(t, want.BaseURL, got.BaseURL)
	assert.Equal(t, want.DatabaseDSN, got.DatabaseDSN)
	assert.Equal(t, want.ShutdownTimeout, got.ShutdownTimeout)
	assert.Equal(t, want.ReadHeaderTimeout, got.ReadHeaderTimeout)
	assert.Equal(t, want.ReadTimeout, got.ReadTimeout)
	assert.Equal(t, want.WriteTimeout, got.WriteTimeout)
	assert.Equal(t, want.IdleTimeout, got.IdleTimeout)
	assert.Equal(t, want.PostgresConnectTimeout, got.PostgresConnectTimeout)
}
