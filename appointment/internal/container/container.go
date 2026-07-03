package container

import (
	"context"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/config"
	pgrepo "github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/postgres"
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
func (c *Container) AppService() *application.Service {
	return application.NewService(pgrepo.NewRepository(c.DB), c.Config.Reminder.TriggerBefore)
}
func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}
	if c.Log != nil {
		_ = c.Log.Sync()
	}
}
func NewMigrator(dsn string) (*migrate.Migrate, error) {
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	sqlDB := stdlib.OpenDB(*db.Config().ConnConfig.Copy())
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		sqlDB.Close()
		db.Close()
		return nil, err
	}
	return migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
}

var _ *sql.DB
