package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	identity "github.com/petretiandrea/beaesthetic-backend/core-contracts/identity"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/cmd/di"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/infra/seed"
	grpcport "github.com/petretiandrea/beaesthetic-backend/identity-service/internal/port/grpc/server"
	httpport "github.com/petretiandrea/beaesthetic-backend/identity-service/internal/port/http/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	if err := root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func root() *cobra.Command {
	root := &cobra.Command{Use: "identity", SilenceUsage: true}
	root.AddCommand(&cobra.Command{Use: "app", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error { return run(cmd.Context()) }})
	root.AddCommand(&cobra.Command{Use: "migrate [up|down|version]", Args: cobra.ExactArgs(1), RunE: migrateCommand})
	root.AddCommand(&cobra.Command{Use: "bootstrap <yaml>", Args: cobra.ExactArgs(1), RunE: bootstrapCommand})
	root.AddCommand(&cobra.Command{Use: "seed-roles [yaml]", Args: cobra.MaximumNArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		path := "seeds/roles.yaml"
		if len(args) == 1 {
			path = args[0]
		}
		c, err := di.New(cmd.Context())
		if err != nil {
			return err
		}
		defer c.GetPostgres().Close()
		return seed.Apply(cmd.Context(), c.GetPostgres(), path)
	}})
	return root
}

func bootstrapCommand(cmd *cobra.Command, args []string) error {
	c, err := di.New(cmd.Context())
	if err != nil {
		return err
	}
	defer c.GetPostgres().Close()
	defer c.Log.Sync()
	return seed.ApplyBootstrap(cmd.Context(), c.GetPostgres(), args[0])
}

func migrateCommand(cmd *cobra.Command, args []string) error {
	c, err := di.New(cmd.Context())
	if err != nil {
		return err
	}
	defer c.GetPostgres().Close()
	defer c.Log.Sync()
	m := c.GetMigrator()
	defer m.Close()
	switch args[0] {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case "down":
		if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "version=%d dirty=%v\n", version, dirty)
	default:
		return fmt.Errorf("unsupported migration command %q", args[0])
	}
	return nil
}
func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	c, err := di.New(ctx)
	if err != nil {
		return err
	}
	defer c.GetPostgres().Close()
	defer c.Log.Sync()
	httpServer := &http.Server{Addr: c.Config.HTTP.Addr, Handler: httpport.New(httpport.Config{InternalAPIKey: c.Config.InternalAPIKey}, c.GetIssuer(), c.GetExchange(), c.GetMemberships()), ReadHeaderTimeout: 5 * time.Second}
	listener, err := net.Listen("tcp", c.Config.GRPC.Addr)
	if err != nil {
		return err
	}
	c.Log.Info("starting identity servers", zap.String("http_addr", c.Config.HTTP.Addr), zap.String("grpc_addr", c.Config.GRPC.Addr))
	grpcServer := grpc.NewServer()
	identity.RegisterTokenServiceServer(grpcServer, grpcport.NewTokenServer(c.Config.InternalAPIKey, c.GetIssuer(), c.GetExchange()))
	identity.RegisterMembershipServiceServer(grpcServer, grpcport.NewMembershipServer(c.Config.InternalAPIKey, c.GetMemberships()))
	errCh := make(chan error, 2)
	go func() { errCh <- httpServer.ListenAndServe() }()
	go func() { errCh <- grpcServer.Serve(listener) }()
	select {
	case <-ctx.Done():
		grpcServer.GracefulStop()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("run identity: %w", err)
	}
}
