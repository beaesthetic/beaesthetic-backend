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
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/config"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/container"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/backfill"
	httpport "github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewRootCommand() *cobra.Command {
	var envFile string
	root := &cobra.Command{Use: "appointment", Short: "Appointment service", SilenceUsage: true}
	root.PersistentFlags().StringVar(&envFile, "env-file", "", "optional dotenv file")
	root.AddCommand(appCommand(&envFile), migrateCommand(&envFile), backfillCommand(&envFile))
	return root
}

func appCommand(envFile *string) *cobra.Command {
	return &cobra.Command{Use: "app", Short: "Start HTTP API", RunE: func(cmd *cobra.Command, args []string) error {
		c, err := container.Build(cmd.Context(), *envFile)
		if err != nil {
			return err
		}
		defer c.Close()
		router := httpport.NewRouter(httpport.NewServer(c.AppService(), c.Log))
		srv := &nethttp.Server{Addr: c.Config.HTTP.Addr, Handler: router}
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		errCh := make(chan error, 1)
		go func() {
			c.Log.Info("starting http server", zap.String("addr", c.Config.HTTP.Addr))
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
				errCh <- err
			}
		}()
		select {
		case <-ctx.Done():
		case err := <-errCh:
			return err
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}}
}

func migrateCommand(envFile *string) *cobra.Command {
	return &cobra.Command{Use: "migrate [up|down|version]", Short: "Run Postgres migrations", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(*envFile)
		if err != nil {
			return err
		}
		m, err := container.NewMigrator(cfg.Postgres.DSN)
		if err != nil {
			return err
		}
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
			v, d, err := m.Version()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "version=%d dirty=%v\n", v, d)
		default:
			return fmt.Errorf("unsupported migration command %q", args[0])
		}
		return nil
	}}
}

func backfillCommand(envFile *string) *cobra.Command {
	return &cobra.Command{Use: "backfill", Short: "Copy legacy Mongo data to Postgres", RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(*envFile)
		if err != nil {
			return err
		}
		log, _ := zap.NewProduction()
		defer log.Sync()
		return backfill.Run(cmd.Context(), cfg, log)
	}}
}
