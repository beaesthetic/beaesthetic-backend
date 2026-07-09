package di

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	postgresinfra "github.com/petretiandrea/beaesthetic-backend/customer/internal/infra/postgres"
)

func (d *DiContainer) GetPostgresDatabase() *sql.DB {
	return singletonWithError(d, "postgresDatabase", func() (*sql.DB, error) {
		db, err := sql.Open("postgres", d.Config.Postgres.DSN)
		if err != nil {
			return nil, err
		}
		return db, db.Ping()
	})
}

func (d *DiContainer) GetMigrator() *migrate.Migrate {
	return singletonWithError(d, "migrator", func() (*migrate.Migrate, error) {
		db := d.GetPostgresDatabase()
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		if err != nil {
			return nil, err
		}
		return migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	})
}

func (d *DiContainer) GetCustomerRepository() *postgresinfra.CustomerRepository {
	return singleton(d, "customerRepository", func() *postgresinfra.CustomerRepository {
		return postgresinfra.NewCustomerRepository(d.GetPostgresDatabase())
	})
}

func (d *DiContainer) GetFidelityRepository() *postgresinfra.FidelityRepository {
	return singleton(d, "fidelityRepository", func() *postgresinfra.FidelityRepository {
		return postgresinfra.NewFidelityRepository(d.GetPostgresDatabase())
	})
}

func (d *DiContainer) GetWalletRepository() *postgresinfra.WalletRepository {
	return singleton(d, "walletRepository", func() *postgresinfra.WalletRepository {
		return postgresinfra.NewWalletRepository(d.GetPostgresDatabase())
	})
}
