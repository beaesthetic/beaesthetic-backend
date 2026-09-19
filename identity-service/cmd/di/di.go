package di

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/config"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/token"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/infra/oidc"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/infra/postgres"
	"go.uber.org/zap"
)

type Container struct {
	Config config.Config
	Log    *zap.Logger
	deps   sync.Map
}

func New(context.Context) (*Container, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	log, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &Container{Config: cfg, Log: log}, nil
}
func singleton[T any](c *Container, key string, factory func() T) T {
	if v, ok := c.deps.Load(key); ok {
		return v.(T)
	}
	v := factory()
	c.deps.Store(key, v)
	return v
}
func singletonError[T any](c *Container, key string, factory func() (T, error)) T {
	if v, ok := c.deps.Load(key); ok {
		return v.(T)
	}
	v, err := factory()
	if err != nil {
		panic(err)
	}
	c.deps.Store(key, v)
	return v
}
func (c *Container) GetPostgres() *pgxpool.Pool {
	return singletonError(c, "postgres", func() (*pgxpool.Pool, error) { return pgxpool.New(context.Background(), c.Config.PostgresDSN) })
}

func (c *Container) GetMigrator() *migrate.Migrate {
	return singletonError(c, "migrator", func() (*migrate.Migrate, error) {
		db := c.GetPostgres()
		sqlDB := stdlib.OpenDB(*db.Config().ConnConfig.Copy())
		driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
		if err != nil {
			sqlDB.Close()
			return nil, err
		}
		return migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	})
}

func stdlibOpenDBFromPool(pool *pgxpool.Pool) *sql.DB {
	return stdlib.OpenDB(*pool.Config().ConnConfig.Copy())
}
func (c *Container) GetIssuer() *token.Issuer {
	return singletonError(c, "issuer", func() (*token.Issuer, error) {
		return token.NewIssuer(token.Config{Issuer: c.Config.Token.Issuer, ActiveKeyID: c.Config.Token.ActiveKeyID, PrivateKeyB64: c.Config.Token.PrivateKeyB64, VerifyKeys: c.Config.Token.VerifyKeys, TTL: c.Config.Token.TTL, Clock: time.Now})
	})
}
func (c *Container) GetMemberships() *postgres.MembershipRepository {
	return singleton(c, "memberships", func() *postgres.MembershipRepository { return postgres.NewMembershipRepository(c.GetPostgres()) })
}
func (c *Container) GetExchange() *application.ExchangeService {
	return singleton(c, "exchange", func() *application.ExchangeService {
		authenticators := make([]application.Authenticator, 0, len(c.Config.OIDC.Providers))
		for _, provider := range c.Config.OIDC.Providers {
			verifier, err := oidc.NewVerifier(context.Background(), provider.Name, provider.Issuer, provider.Audience, provider.DiscoveryURL)
			if err != nil {
				panic(err)
			}
			authenticators = append(authenticators, verifier)
		}
		return application.NewExchangeService(c.GetIssuer(), authenticators, []application.Authorizer{application.NewMembershipAuthorizer(c.GetMemberships())})
	})
}
