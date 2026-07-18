package di

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (d *DiContainer) GetPostgresDatabase() *pgxpool.Pool {
	// pgxpool
	return singletonWithError(d, "postgresDatabase", func() (*pgxpool.Pool, error) {
		return pgxpool.New(context.Background(), d.Config.Postgres.DSN)
	})
}
