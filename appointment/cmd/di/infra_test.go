package di

import (
	"sync"
	"testing"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/config"
	"go.uber.org/zap"
)

func TestRiverRuntimeClientCanBeBuiltWithoutDependencyCycle(t *testing.T) {
	container := &DiContainer{
		Config: config.Config{
			Postgres: config.PostgresConfig{DSN: "postgres://localhost/appointment?sslmode=disable"},
		},
		Log:  zap.NewNop(),
		deps: &sync.Map{},
	}
	t.Cleanup(func() { container.GetPostgresDatabase().Close() })

	runtimeClient := container.GetRiverClient()
	insertClient := container.GetRiverInsertClient()

	if runtimeClient == nil || insertClient == nil {
		t.Fatal("expected both River clients to be constructed")
	}
	if runtimeClient == insertClient {
		t.Fatal("runtime and transactional insert clients must be separate")
	}
}
