package config

import (
	"testing"
	"time"
)

func TestLoadMapsEnvironmentVariables(t *testing.T) {
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("SCHEDULER_NAME", "test-scheduler")
	t.Setenv("SCHEDULER_POLLING_INTERVAL", "5s")
	t.Setenv("SCHEDULER_PEEK_LEASE_TTL", "7s")
	t.Setenv("SCHEDULER_PEEK_BATCH_SIZE", "12")
	t.Setenv("POSTGRES_DSN", "postgres://example")
	t.Setenv("RABBIT_HOST", "rabbit")
	t.Setenv("RABBIT_USERNAME", "user")
	t.Setenv("RABBIT_EXCHANGE", "scheduler-exchange")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Fatalf("Server.Port = %d", cfg.Server.Port)
	}
	if cfg.Scheduler.Name != "test-scheduler" {
		t.Fatalf("Scheduler.Name = %q", cfg.Scheduler.Name)
	}
	if cfg.Scheduler.PollingInterval != 5*time.Second {
		t.Fatalf("Scheduler.PollingInterval = %s", cfg.Scheduler.PollingInterval)
	}
	if cfg.Scheduler.PeekLeaseTTL != 7*time.Second {
		t.Fatalf("Scheduler.PeekLeaseTTL = %s", cfg.Scheduler.PeekLeaseTTL)
	}
	if cfg.Scheduler.PeekBatchSize != 12 {
		t.Fatalf("Scheduler.PeekBatchSize = %d", cfg.Scheduler.PeekBatchSize)
	}
	if cfg.Postgres.DSN != "postgres://example" {
		t.Fatalf("Postgres.DSN = %q", cfg.Postgres.DSN)
	}
	if cfg.RabbitMQ.Host != "rabbit" {
		t.Fatalf("RabbitMQ.Host = %q", cfg.RabbitMQ.Host)
	}
	if cfg.RabbitMQ.Username != "user" {
		t.Fatalf("RabbitMQ.Username = %q", cfg.RabbitMQ.Username)
	}
	if cfg.RabbitMQ.Exchange != "scheduler-exchange" {
		t.Fatalf("RabbitMQ.Exchange = %q", cfg.RabbitMQ.Exchange)
	}
}
