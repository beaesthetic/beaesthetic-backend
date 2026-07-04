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
	"github.com/petretiandrea/beaesthetic-backend/appointment/cmd/di"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/backfill"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		c, err := di.NewDiContainer(ctx, *envFile)
		if err != nil {
			return err
		}
		defer func() {
			c.GetPostgresDatabase().Close()
			_ = c.Log.Sync()
		}()

		httpServer := c.GetHttpServer()
		schedulerConsumer := c.GetSchedulerQueueConsumer()
		notificationConfirmConsumer := c.GetNotificationConfirmQueueConsumer()

		group, groupCtx := errgroup.WithContext(ctx)

		group.Go(func() error {
			c.Log.Info("starting http server", zap.String("addr", c.Config.HTTP.Addr))
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
				return fmt.Errorf("run http server: %w", err)
			}
			return nil
		})
		group.Go(func() error {
			<-groupCtx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("shutdown http server: %w", err)
			}
			return nil
		})
		group.Go(func() error {
			c.Log.Info("starting scheduler queue consumer", zap.String("queue", c.Config.RabbitMQ.SchedulerQueue))
			if err := schedulerConsumer.Run(groupCtx); err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("run scheduler queue consumer: %w", err)
			}
			return nil
		})
		group.Go(func() error {
			c.Log.Info("starting notification confirm queue consumer", zap.String("queue", c.Config.RabbitMQ.NotificationConfirmQueue))
			if err := notificationConfirmConsumer.Run(groupCtx); err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("run notification confirm queue consumer: %w", err)
			}
			return nil
		})

		return group.Wait()
	}}
}

func migrateCommand(envFile *string) *cobra.Command {
	return &cobra.Command{Use: "migrate [up|down|version]", Short: "Run Postgres migrations", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		c, err := di.NewDiContainer(cmd.Context(), *envFile)
		if err != nil {
			return err
		}
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
		c, err := di.NewDiContainer(cmd.Context(), *envFile)
		if err != nil {
			return err
		}
		defer func() { _ = c.Log.Sync() }()
		return backfill.Run(cmd.Context(), c.Config, c.Log)
	}}
}
