package container

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/beaesthetic/scheduler/internal/application"
	"github.com/beaesthetic/scheduler/internal/config"
	"github.com/beaesthetic/scheduler/internal/infra/postgres"
	"github.com/beaesthetic/scheduler/internal/infra/rabbitmq"
	httpport "github.com/beaesthetic/scheduler/internal/port/http"
	schedruntime "github.com/beaesthetic/scheduler/internal/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Container struct {
	Config          *config.Config
	Logger          *zap.Logger
	DB              *pgxpool.Pool
	Publisher       *rabbitmq.Publisher
	Store           *postgres.JobRepository
	Scheduler       *application.SchedulerService
	Runtime         *schedruntime.Runtime
	HTTP            *http.Server
	tracerProvider  *sdktrace.TracerProvider
	shutdownTimeout time.Duration
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	logger, err := NewLogger(cfg)
	if err != nil {
		return nil, err
	}

	tp, err := newTracerProvider(ctx, cfg)
	if err != nil {
		_ = logger.Sync()
		return nil, err
	}
	if tp != nil {
		otel.SetTracerProvider(tp)
	}

	db, err := postgres.Open(ctx, cfg.Postgres.DSN)
	if err != nil {
		_ = logger.Sync()
		return nil, err
	}

	publisher, err := rabbitmq.NewPublisher(cfg.RabbitMQ)
	if err != nil {
		db.Close()
		_ = logger.Sync()
		return nil, err
	}

	store := postgres.NewJobRepository(db, cfg.Scheduler.Name)
	scheduler := application.NewSchedulerService(store)
	runtime := schedruntime.New(store, publisher, logger, cfg.Scheduler.PollingInterval, cfg.Scheduler.PeekLeaseTTL, cfg.Scheduler.PeekBatchSize)
	router := httpport.NewRouter(scheduler, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router.Engine(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Container{
		Config:          cfg,
		Logger:          logger,
		DB:              db,
		Publisher:       publisher,
		Store:           store,
		Scheduler:       scheduler,
		Runtime:         runtime,
		HTTP:            server,
		tracerProvider:  tp,
		shutdownTimeout: 30 * time.Second,
	}, nil
}

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.Set(cfg.Log.Level); err != nil {
		return nil, err
	}

	if cfg.Log.Development {
		zapCfg := zap.NewDevelopmentConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(level)
		return zapCfg.Build()
	}

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	return zapCfg.Build()
}

func (c *Container) Run(ctx context.Context) error {
	errs := make(chan error, 2)

	go func() {
		c.Logger.Info("starting HTTP server", zap.String("addr", c.HTTP.Addr))
		if err := c.HTTP.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
			return
		}
		errs <- nil
	}()

	go func() {
		c.Logger.Info("starting scheduler runtime",
			zap.String("scheduler_name", c.Config.Scheduler.Name),
			zap.Duration("polling_interval", c.Config.Scheduler.PollingInterval),
			zap.Int("batch_size", c.Config.Scheduler.PeekBatchSize),
		)
		err := c.Runtime.Run(ctx)
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		errs <- err
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errs:
		return err
	}
}

func (c *Container) Close(ctx context.Context) error {
	var result error

	if c.HTTP != nil {
		if err := c.HTTP.Shutdown(ctx); err != nil {
			result = errors.Join(result, err)
		}
	}
	if c.Publisher != nil {
		if err := c.Publisher.Close(); err != nil {
			result = errors.Join(result, err)
		}
	}
	if c.DB != nil {
		c.DB.Close()
	}
	if c.tracerProvider != nil {
		if err := c.tracerProvider.Shutdown(ctx); err != nil {
			result = errors.Join(result, err)
		}
	}
	if c.Logger != nil {
		if err := c.Logger.Sync(); err != nil {
			result = errors.Join(result, err)
		}
	}

	return result
}

func newTracerProvider(ctx context.Context, cfg *config.Config) (*sdktrace.TracerProvider, error) {
	if !cfg.Otel.Enabled || cfg.Otel.Endpoint == "" {
		return nil, nil
	}

	endpoint := strings.TrimPrefix(cfg.Otel.Endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimSuffix(endpoint, "/")
	if _, _, err := net.SplitHostPort(endpoint); err != nil {
		endpoint = endpoint + ":4317"
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	), nil
}
