package config

import "testing"

func TestLoadReadsEnvPrefix(t *testing.T) {
	t.Setenv("ENV_APP_ENV", "test")
	t.Setenv("ENV_HTTP_ADDR", ":9090")
	t.Setenv("ENV_POSTGRES_DSN", "postgres://user:pass@localhost:5432/customers")

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
}
