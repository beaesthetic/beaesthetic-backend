package container

import (
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func stdlibOpenDBFromPool(pool *pgxpool.Pool) *sql.DB {
	config := pool.Config().ConnConfig.Copy()
	return stdlib.OpenDB(*config)
}
