package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionContextKey struct{}

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type ContextDB struct {
	db DBTX
}

func NewContextDB(db DBTX) *ContextDB {
	return &ContextDB{db: db}
}

func (db *ContextDB) Tx(ctx context.Context, atomicFn func(ctx context.Context) error) (err error) {
	pool, ok := db.db.(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("postgres context db requires *pgxpool.Pool to begin transaction, got %T", db.db)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = atomicFn(contextWithTx(ctx, tx))
	return err
}

func (db *ContextDB) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return db.executor(ctx).Exec(ctx, sql, arguments...)
}

func (db *ContextDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return db.executor(ctx).Query(ctx, sql, args...)
}

func (db *ContextDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return db.executor(ctx).QueryRow(ctx, sql, args...)
}

func (db *ContextDB) executor(ctx context.Context) DBTX {
	if tx, ok := ctx.Value(transactionContextKey{}).(DBTX); ok && tx != nil {
		return tx
	}
	return db.db
}

func contextWithTx(ctx context.Context, tx DBTX) context.Context {
	return context.WithValue(ctx, transactionContextKey{}, tx)
}
