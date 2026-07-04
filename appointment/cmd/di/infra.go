package di

import (
	"context"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func (d *DiContainer) GetPostgresDatabase() *pgxpool.Pool {
	return singletonWithError(d, "postgresDatabase", func() (*pgxpool.Pool, error) {
		return pgxpool.New(context.Background(), d.Config.Postgres.DSN)
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
