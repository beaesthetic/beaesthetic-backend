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
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/config"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/container"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/backfill"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/messaging"
	httpport "github.com/petretiandrea/beaesthetic-backend/notification/internal/port/http"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewRootCommand() *cobra.Command {
	var envFile string
	root := &cobra.Command{
		Use:          "notification",
		Short:        "Notification service",
		SilenceUsage: true,
	}
	root.PersistentFlags().StringVar(&envFile, "env-file", "", "optional dotenv file")
	root.AddCommand(appCommand(&envFile), migrateCommand(&envFile), backfillCommand(&envFile), rabbitCommand(&envFile))
	return root
}

func appCommand(envFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "app",
		Short: "Start the HTTP API and RabbitMQ consumer",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := container.Build(cmd.Context(), *envFile)
			if err != nil {
				return err
			}
			defer c.Close()

			service := c.NotificationService()
			router := httpport.NewRouter(httpport.NewServer(service, c.Log))
			server := &nethttp.Server{Addr: c.Config.HTTP.Addr, Handler: router}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			errCh := make(chan error, 2)
			go func() {
				c.Log.Info("starting http server", zap.String("addr", c.Config.HTTP.Addr))
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
					errCh <- err
				}
			}()
			go func() {
				errCh <- messaging.NewConsumer(c.Config.RabbitMQ, service, c.Log).Run(ctx)
			}()

			select {
			case <-ctx.Done():
			case err := <-errCh:
				if err != nil && !errors.Is(err, context.Canceled) {
					return err
				}
			}
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			return server.Shutdown(shutdownCtx)
		},
	}
}

func migrateCommand(envFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate [up|down|version]",
		Short: "Run PostgreSQL migrations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
				version, dirty, err := m.Version()
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "version=%d dirty=%v\n", version, dirty)
			default:
				return fmt.Errorf("unsupported migration command %q", args[0])
			}
			return nil
		},
	}
}

func backfillCommand(envFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "backfill",
		Short: "Copy legacy MongoDB notifications into PostgreSQL",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*envFile)
			if err != nil {
				return err
			}
			log, _ := zap.NewProduction()
			defer log.Sync()
			return backfill.Run(cmd.Context(), cfg, log)
		},
	}
}

func rabbitCommand(envFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "rabbitmq-topology",
		Short: "Apply RabbitMQ queue topology",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*envFile)
			if err != nil {
				return err
			}
			return messaging.ApplyTopology(cfg.RabbitMQ)
		},
	}
}
