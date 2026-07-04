package config

import (
	"testing"
	"time"
)

func TestLoadEnvironment(t *testing.T) {
	t.Setenv("ENV_POSTGRES_DSN", "postgres://test")
	t.Setenv("ENV_REMINDER_TRIGGER__BEFORE", "2h")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Postgres.DSN != "postgres://test" {
		t.Fatalf("dsn=%q", cfg.Postgres.DSN)
	}
	if cfg.Reminder.TriggerBefore != 2*time.Hour {
		t.Fatalf("trigger=%s", cfg.Reminder.TriggerBefore)
	}
}
