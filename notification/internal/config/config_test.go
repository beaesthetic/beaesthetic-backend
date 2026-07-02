package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Name != "beaesthetic-notifications" {
		t.Fatalf("App.Name = %q", cfg.App.Name)
	}
	if cfg.HTTP.Addr != ":8080" {
		t.Fatalf("HTTP.Addr = %q", cfg.HTTP.Addr)
	}
	if cfg.RabbitMQ.RetryTTL != time.Millisecond {
		t.Fatalf("RetryTTL = %s, want 1ms", cfg.RabbitMQ.RetryTTL)
	}
}

func TestLoadOverridesFromEnvironment(t *testing.T) {
	t.Setenv("POSTGRES__DSN", "postgres://user:pass@postgres:5432/notifications")
	t.Setenv("RABBITMQ__URL", "amqp://user:pass@rabbitmq:5672/")
	t.Setenv("SMS_GATEWAY__API_KEY", "secret")
	t.Setenv("RABBITMQ__RETRY_TTL", "5s")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Postgres.DSN != "postgres://user:pass@postgres:5432/notifications" {
		t.Fatalf("Postgres.DSN = %q", cfg.Postgres.DSN)
	}
	if cfg.RabbitMQ.URL != "amqp://user:pass@rabbitmq:5672/" {
		t.Fatalf("RabbitMQ.URL = %q", cfg.RabbitMQ.URL)
	}
	if cfg.SMSGateway.APIKey != "secret" {
		t.Fatalf("SMSGateway.APIKey = %q", cfg.SMSGateway.APIKey)
	}
	if cfg.RabbitMQ.RetryTTL != 5*time.Second {
		t.Fatalf("RetryTTL = %s, want 5s", cfg.RabbitMQ.RetryTTL)
	}
}

func TestLoadOverridesFromDotEnvFile(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	content := []byte("POSTGRES__DSN=postgres://file\nMONGO__DATABASE=notifications-file\n")
	if err := os.WriteFile(envFile, content, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(envFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Postgres.DSN != "postgres://file" {
		t.Fatalf("Postgres.DSN = %q", cfg.Postgres.DSN)
	}
	if cfg.Mongo.Database != "notifications-file" {
		t.Fatalf("Mongo.Database = %q", cfg.Mongo.Database)
	}
}
