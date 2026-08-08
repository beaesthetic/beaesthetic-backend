package di

import (
	"context"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/jobs"
	app_postgres "github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/postgres"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
	outboxpostgres "github.com/petretiandrea/outbox-go/pkg/outbox/postgres"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
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
		driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
		if err != nil {
			sqlDB.Close()
			db.Close()
			return nil, err
		}
		return migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	})
}

func (d *DiContainer) GetRiverClient() *river.Client[pgx.Tx] {
	return singletonWithError(d, "riverClient", func() (*river.Client[pgx.Tx], error) {
		riverConfig := d.GetRiverReminderConfig()
		workers := river.NewWorkers()
		if err := river.AddWorkerSafely(workers, jobs.NewSendAppointmentReminderWorker(d.GetAppointmentLifecycleServiceV2())); err != nil {
			return nil, err
		}
		return river.NewClient(riverpgxv5.New(d.GetPostgresDatabase()), &river.Config{
			Queues: map[string]river.QueueConfig{
				riverConfig.Queue: {MaxWorkers: riverConfig.Workers},
			},
			Workers: workers,
		})
	})
}

func (d *DiContainer) GetRiverInsertClient() *river.Client[pgx.Tx] {
	return singletonWithError(d, "riverInsertClient", func() (*river.Client[pgx.Tx], error) {
		return river.NewClient(riverpgxv5.New(d.GetPostgresDatabase()), &river.Config{})
	})
}

func (d *DiContainer) MigrateRiver(ctx context.Context) error {
	migrator, err := rivermigrate.New(riverpgxv5.New(d.GetPostgresDatabase()), nil)
	if err != nil {
		return err
	}
	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	return err
}

func (d *DiContainer) GetRiverJobInserter() *app_postgres.RiverJobInserter {
	return singleton(d, "riverJobInserter", func() *app_postgres.RiverJobInserter {
		return app_postgres.NewRiverJobInserter(d.GetPostgresContextDB(), d.GetRiverInsertClient())
	})
}

func stdlibOpenDBFromPool(pool *pgxpool.Pool) *sql.DB {
	config := pool.Config().ConnConfig.Copy()
	return stdlib.OpenDB(*config)
}
