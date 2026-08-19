package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithoutEnvironmentUsesZeroValues(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Name != "" {
		t.Fatalf("App.Name = %q", cfg.App.Name)
	}
	if cfg.HTTP.Addr != "" {
		t.Fatalf("HTTP.Addr = %q", cfg.HTTP.Addr)
	}
}

func TestLoadOverridesFromEnvironment(t *testing.T) {
	t.Setenv("POSTGRES__DSN", "postgres://user:pass@postgres:5432/notifications")
	t.Setenv("RABBITMQ__URL", "amqp://user:pass@rabbitmq:5672/")
	t.Setenv("SMS_GATEWAY__API_KEY", "secret")

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
}

func TestLoadOverridesFromDotEnvFile(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	content := []byte("POSTGRES__DSN=postgres://file\nRABBITMQ__URL=amqp://file\n")
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
	if cfg.RabbitMQ.URL != "amqp://file" {
		t.Fatalf("RabbitMQ.URL = %q", cfg.RabbitMQ.URL)
	}
}
