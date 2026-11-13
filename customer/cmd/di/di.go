package di

import (
	"context"
	"sync"

	"github.com/petretiandrea/beaesthetic-backend/core/auth"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/config"
	"go.uber.org/zap"
)

type DiContainer struct {
	Config config.Config
	Log    *zap.Logger
	deps   *sync.Map
}

func NewDiContainer(ctx context.Context, envFile string) (*DiContainer, error) {
	cfg, err := config.Load(envFile)
	if err != nil {
		return nil, err
	}
	log, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &DiContainer{Config: cfg, Log: log, deps: &sync.Map{}}, nil
}

func singleton[T any](container *DiContainer, key string, factory func() T) T {
	if dep, ok := container.deps.Load(key); ok {
		return dep.(T)
	}
	instance := factory()
	container.deps.Store(key, instance)
	return instance
}

func singletonWithError[T any](container *DiContainer, key string, factory func() (T, error)) T {
	if dep, ok := container.deps.Load(key); ok {
		return dep.(T)
	}
	instance, err := factory()
	if err != nil {
		container.Log.Error("failed to create singleton instance", zap.String("key", key), zap.Error(err))
		panic(err)
	}
	container.deps.Store(key, instance)
	return instance
}

func singletonWithErrorResult[T any](container *DiContainer, key string, factory func() (T, error)) (T, error) {
	if dep, ok := container.deps.Load(key); ok {
		return dep.(T), nil
	}
	instance, err := factory()
	if err != nil {
		var zero T
		return zero, err
	}
	actual, _ := container.deps.LoadOrStore(key, instance)
	return actual.(T), nil
}

func (d *DiContainer) GetAuthVerifier() (*auth.Verifier, error) {
	return singletonWithErrorResult(d, "authVerifier", d.Config.Auth.NewVerifier)
}
