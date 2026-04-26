package config

import (
	"strings"
	"testing"
	"time"
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
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errSubstr != "" && !strings.Contains(err.Error(), tc.errSubstr) {
					t.Fatalf("expected error to contain %q, got %q", tc.errSubstr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

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
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errSubstr != "" && !strings.Contains(err.Error(), tc.errSubstr) {
					t.Fatalf("expected error to contain %q, got %q", tc.errSubstr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
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

	if got.AppEnv != want.AppEnv {
		t.Fatalf("expected AppEnv %q, got %q", want.AppEnv, got.AppEnv)
	}
	if got.StorageType != want.StorageType {
		t.Fatalf("expected StorageType %q, got %q", want.StorageType, got.StorageType)
	}
	if got.HTTPAddr != want.HTTPAddr {
		t.Fatalf("expected HTTPAddr %q, got %q", want.HTTPAddr, got.HTTPAddr)
	}
	if got.BaseURL != want.BaseURL {
		t.Fatalf("expected BaseURL %q, got %q", want.BaseURL, got.BaseURL)
	}
	if got.DatabaseDSN != want.DatabaseDSN {
		t.Fatalf("expected DatabaseDSN %q, got %q", want.DatabaseDSN, got.DatabaseDSN)
	}
	if got.ShutdownTimeout != want.ShutdownTimeout {
		t.Fatalf("expected ShutdownTimeout %s, got %s", want.ShutdownTimeout, got.ShutdownTimeout)
	}
	if got.ReadHeaderTimeout != want.ReadHeaderTimeout {
		t.Fatalf("expected ReadHeaderTimeout %s, got %s", want.ReadHeaderTimeout, got.ReadHeaderTimeout)
	}
	if got.ReadTimeout != want.ReadTimeout {
		t.Fatalf("expected ReadTimeout %s, got %s", want.ReadTimeout, got.ReadTimeout)
	}
	if got.WriteTimeout != want.WriteTimeout {
		t.Fatalf("expected WriteTimeout %s, got %s", want.WriteTimeout, got.WriteTimeout)
	}
	if got.IdleTimeout != want.IdleTimeout {
		t.Fatalf("expected IdleTimeout %s, got %s", want.IdleTimeout, got.IdleTimeout)
	}
	if got.PostgresConnectTimeout != want.PostgresConnectTimeout {
		t.Fatalf("expected PostgresConnectTimeout %s, got %s", want.PostgresConnectTimeout, got.PostgresConnectTimeout)
	}
}
