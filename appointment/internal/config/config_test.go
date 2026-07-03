package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadEnvironment(t *testing.T) {
	t.Setenv("POSTGRES__DSN", "postgres://test")
	t.Setenv("REMINDER__TRIGGER_BEFORE", "2h")
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
func TestLoadDotEnv(t *testing.T) {
	file := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(file, []byte("HTTP__ADDR=:9090\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("addr=%q", cfg.HTTP.Addr)
	}
}
