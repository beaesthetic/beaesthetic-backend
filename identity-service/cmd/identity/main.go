package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/config"
	identityhttp "github.com/petretiandrea/beaesthetic-backend/identity-service/internal/http"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/membership"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/oauth"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/token"
)

func main() {
	if err := run(); err != nil {
		slog.Error("identity service stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	issuer, err := token.NewIssuer(token.Config{
		Issuer:        cfg.Token.Issuer,
		ActiveKeyID:   cfg.Token.ActiveKeyID,
		PrivateKeyB64: cfg.Token.PrivateKeyB64,
		VerifyKeys:    cfg.Token.VerifyKeys,
		TTL:           cfg.Token.TTL,
		Clock:         time.Now,
	})
	if err != nil {
		return fmt.Errorf("configure token issuer: %w", err)
	}
	pool, err := pgxpool.New(context.Background(), cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("connect PostgreSQL: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}
	exchange := oauth.NewExchangeService(issuer, []oauth.Authenticator{oauth.NewFirebaseVerifier(cfg.FirebaseProjectID, nil)}, []oauth.Authorizer{oauth.NewMembershipAuthorizer(membership.NewPostgresRepository(pool))})

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           identityhttp.NewHandler(identityhttp.Config{InternalAPIKey: cfg.InternalAPIKey}, issuer, exchange),
		ReadHeaderTimeout: 5 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() {
		slog.Info("starting identity service", "addr", cfg.HTTP.Addr, "issuer", cfg.Token.Issuer)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("run http server: %w", err)
	}
}
