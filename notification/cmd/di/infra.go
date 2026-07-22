package di

import (
	"context"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	app_postgres "github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/postgres"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
	outboxpostgres "github.com/petretiandrea/outbox-go/pkg/outbox/postgres"
)

func (d *DiContainer) GetPostgresDatabase() *pgxpool.Pool {
	return singletonWithError(d, "postgresDatabase", func() (*pgxpool.Pool, error) {
		return pgxpool.New(context.Background(), d.Config.Postgres.DSN)
	})
}

func (d *DiContainer) GetPostgresContextDB() *app_postgres.ContextDB {
	return singleton(d, "postgresContextDB", func() *app_postgres.ContextDB {
		return app_postgres.NewContextDB(d.GetPostgresDatabase())
	})
}

func (d *DiContainer) GetOutboxPublisher() outbox.Publisher {
	return singletonWithError(d, "outboxPublisher", func() (outbox.Publisher, error) {
		return outboxpostgres.NewPublisher(d.GetPostgresContextDB(), outboxpostgres.PublisherConfig{
			TableName: "outbox_messages",
		})
	})
}

func (d *DiContainer) GetMigrator() *migrate.Migrate {
	return singletonWithError(d, "migrator", func() (*migrate.Migrate, error) {
		db := d.GetPostgresDatabase()
		sqlDB := stdlibOpenDBFromPool(db)
		driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
		if err != nil {
			sqlDB.Close()
			db.Close()
			return nil, err
		}
		return migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	})
}

func stdlibOpenDBFromPool(pool *pgxpool.Pool) *sql.DB {
	config := pool.Config().ConnConfig.Copy()
	return stdlib.OpenDB(*config)
}
