package config

import (
	"encoding/base64"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
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

func TestAuthConfiguration(t *testing.T) {
	key := paseto.NewV4AsymmetricSecretKey()
	encoded := base64.RawURLEncoding.EncodeToString(key.Public().ExportBytes())
	t.Setenv("ENV_AUTH_ISSUER", "identity")
	t.Setenv("ENV_AUTH_PUBLIC__KEYS", `{"key.with_dots":"`+encoded+`"}`)
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.Auth.NewVerifier(); err != nil {
		t.Fatal(err)
	}
	for _, c := range []auth.Config{
		{}, {Issuer: "identity", PublicKeysJSON: "{}"},
		{Issuer: "identity", PublicKeysJSON: "bad"},
		{Issuer: "identity", PublicKeysJSON: `{"key":"bad"}`},
		{PublicKeysJSON: cfg.Auth.PublicKeysJSON},
	} {
		if _, err := c.NewVerifier(); err == nil {
			t.Errorf("accepted invalid auth config: %+v", c)
		}
	}
}
