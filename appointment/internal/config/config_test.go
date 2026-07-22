package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadEnvironment(t *testing.T) {
	t.Setenv("ENV_POSTGRES_DSN", "postgres://test")
	t.Setenv("ENV_REMINDER_TRIGGER__BEFORE", "2h")
	t.Setenv("ENV_REMINDER_SCHEDULER__PROVIDER", "river")
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
	if cfg.Reminder.SchedulerProvider != "river" {
		t.Fatalf("scheduler provider=%q", cfg.Reminder.SchedulerProvider)
	}
}

func TestLoadEnvFile(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(
		envFile,
		[]byte("ENV_POSTGRES_DSN=postgres://file\nENV_REMINDER_TRIGGER__BEFORE=45m\n"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(envFile)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Postgres.DSN != "postgres://file" {
		t.Fatalf("dsn=%q", cfg.Postgres.DSN)
	}
	if cfg.Reminder.TriggerBefore != 45*time.Minute {
		t.Fatalf("trigger=%s", cfg.Reminder.TriggerBefore)
	}
}

func TestEnvironmentOverridesEnvFile(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envFile, []byte("ENV_POSTGRES_DSN=postgres://file\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ENV_POSTGRES_DSN", "postgres://env")

	cfg, err := Load(envFile)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Postgres.DSN != "postgres://env" {
		t.Fatalf("dsn=%q", cfg.Postgres.DSN)
	}
}
