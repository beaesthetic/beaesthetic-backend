package config

import (
	"testing"
	"time"
)

func TestLoadReadsEnvPrefix(t *testing.T) {
	t.Setenv("ENV_APP_ENV", "test")
	t.Setenv("ENV_HTTP_ADDR", ":9090")
	t.Setenv("ENV_POSTGRES_DSN", "postgres://user:pass@localhost:5432/customers")
	t.Setenv("ENV_REDIS_URI", "redis://example:6379")
	t.Setenv("ENV_REDIS_CUSTOMERS__TTL", "2h")
	t.Setenv("ENV_REDIS_CUSTOMERS__SEARCH__TTL", "10m")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.App.Env != "test" {
		t.Fatalf("App.Env = %q, want test", cfg.App.Env)
	}
	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("HTTP.Addr = %q, want :9090", cfg.HTTP.Addr)
	}
	if cfg.Postgres.DSN != "postgres://user:pass@localhost:5432/customers" {
		t.Fatalf("Postgres.DSN = %q", cfg.Postgres.DSN)
	}
	if cfg.Redis.URI != "redis://example:6379" {
		t.Fatalf("Redis.URI = %q", cfg.Redis.URI)
	}
	if cfg.Redis.CustomersTTL != 2*time.Hour {
		t.Fatalf("Redis.CustomersTTL = %s", cfg.Redis.CustomersTTL)
	}
	if cfg.Redis.CustomersSearchTTL != 10*time.Minute {
		t.Fatalf("Redis.CustomersSearchTTL = %s", cfg.Redis.CustomersSearchTTL)
	}
}
