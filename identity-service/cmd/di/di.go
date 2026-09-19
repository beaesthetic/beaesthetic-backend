package di

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/config"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/token"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/infra/firebase"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/infra/postgres"
)

type Container struct {
	Config config.Config
	deps   sync.Map
}

func New(context.Context) (*Container, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &Container{Config: cfg}, nil
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
		return application.NewExchangeService(c.GetIssuer(), []application.Authenticator{firebase.NewFirebaseVerifier(c.Config.FirebaseProjectID, nil)}, []application.Authorizer{application.NewMembershipAuthorizer(c.GetMemberships())})
	})
}
