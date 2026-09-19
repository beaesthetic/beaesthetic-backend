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

	identity "github.com/petretiandrea/beaesthetic-backend/core-contracts/identity"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/cmd/di"
	grpcport "github.com/petretiandrea/beaesthetic-backend/identity-service/internal/port/grpc/server"
	httpport "github.com/petretiandrea/beaesthetic-backend/identity-service/internal/port/http/server"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func main() {
	if err := root().Execute(); err != nil {
		os.Exit(1)
	}
}
func root() *cobra.Command {
	return &cobra.Command{Use: "identity", SilenceUsage: true, RunE: func(cmd *cobra.Command, _ []string) error { return run(cmd.Context()) }}
}
func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	c, err := di.New(ctx)
	if err != nil {
		return err
	}
	defer c.GetPostgres().Close()
	httpServer := &http.Server{Addr: c.Config.HTTP.Addr, Handler: httpport.New(httpport.Config{InternalAPIKey: c.Config.InternalAPIKey}, c.GetIssuer(), c.GetExchange(), c.GetMemberships()), ReadHeaderTimeout: 5 * time.Second}
	listener, err := net.Listen("tcp", c.Config.GRPC.Addr)
	if err != nil {
		return err
	}
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
