package container

import (
	"context"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/config"
	postgresrepo "github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/postgres"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/provider"
	"go.uber.org/zap"
)

type Container struct {
	Config config.Config
	Log    *zap.Logger
	DB     *pgxpool.Pool
}

func Build(ctx context.Context, envFile string) (*Container, error) {
	cfg, err := config.Load(envFile)
	if err != nil {
		return nil, err
	}
	log, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	db, err := pgxpool.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		return nil, err
	}
	return &Container{Config: cfg, Log: log, DB: db}, nil
}

func (container *Container) NotificationService() *application.NotificationService {
	repo := postgresrepo.NewNotificationRepository(container.DB)
	sms := provider.NewSMSProvider(container.Config.SMSGateway)
	compound := provider.NewCompoundProvider(sms)
	return application.NewNotificationService(repo, compound)
}

func (container *Container) Close() {
	if container.DB != nil {
		container.DB.Close()
	}
	if container.Log != nil {
		_ = container.Log.Sync()
	}
}

func NewMigrator(dsn string) (*migrate.Migrate, error) {
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	sqlDB := stdlibOpenDBFromPool(db)
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		sqlDB.Close()
		db.Close()
		return nil, err
	}
	return migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
}
