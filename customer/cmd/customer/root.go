package main

import (
	"context"
	"errors"
	"fmt"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/petretiandrea/beaesthetic-backend/customer/cmd/di"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func NewRootCommand() *cobra.Command {
	var envFile string
	root := &cobra.Command{Use: "customer", Short: "Customer service", SilenceUsage: true}
	root.PersistentFlags().StringVar(&envFile, "env-file", "", "optional dotenv file")
	root.AddCommand(appCommand(&envFile), migrateCommand(&envFile))
	return root
}

func appCommand(envFile *string) *cobra.Command {
	return &cobra.Command{Use: "app", Short: "Start the HTTP API", RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		c, err := di.NewDiContainer(ctx, *envFile)
		if err != nil {
			return err
		}
		defer func() { c.GetPostgresDatabase().Close(); _ = c.Log.Sync() }()
		httpServer := c.GetHttpServer()
		serverErr := make(chan error, 1)
		go func() {
			c.Log.Info("starting http server", zap.String("addr", c.Config.HTTP.Addr))
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
				serverErr <- fmt.Errorf("run http server: %w", err)
				return
			}
			serverErr <- nil
		}()

		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("shutdown http server: %w", err)
			}
			return <-serverErr
		case err := <-serverErr:
			return err
		}
	}}
}

func migrateCommand(envFile *string) *cobra.Command {
	return &cobra.Command{Use: "migrate [up|down|version]", Short: "Run PostgreSQL migrations", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		c, err := di.NewDiContainer(cmd.Context(), *envFile)
		if err != nil {
			return err
		}
		defer func() { c.GetPostgresDatabase().Close(); _ = c.Log.Sync() }()
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
	}}
}
