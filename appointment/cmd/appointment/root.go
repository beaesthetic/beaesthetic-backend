package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/petretiandrea/beaesthetic-backend/appointment/cmd/di"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/telemetry"
	appruntime "github.com/petretiandrea/beaesthetic-backend/core-contracts/runtime"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewRootCommand() *cobra.Command {
	var envFile string
	root := &cobra.Command{Use: "appointment", Short: "Appointment service", SilenceUsage: true}
	root.PersistentFlags().StringVar(&envFile, "env-file", "", "optional dotenv file")
	root.AddCommand(appCommand(&envFile), migrateCommand(&envFile))
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
		shutdownTelemetry, err := telemetry.Init(ctx, c.Config.App.Name)
		if err != nil {
			return fmt.Errorf("initialize OpenTelemetry: %w", err)
		}
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := shutdownTelemetry(shutdownCtx); err != nil {
				c.Log.Error("shutdown OpenTelemetry", zap.Error(err))
			}
			c.GetPostgresDatabase().Close()
			_ = c.Log.Sync()
		}()

		runner := appruntime.NewRunner(c.Log)
		riverClient := c.GetRiverClient()
		runner.Add(appruntime.StartStop("river reminder scheduler", riverClient.Start, riverClient.Stop, 10*time.Second))
		runner.Add(appruntime.HTTPServer("http server", c.GetHttpServer(), 10*time.Second))
		runner.Add(appruntime.Consumer("appointment lifecycle consumer", c.GetAppointmentLifecycleConsumer()))
		runner.Add(appruntime.Consumer("notification outcomes consumer", c.GetNotificationOutcomeQueueConsumer()))

		return runner.Run(ctx)
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
			if err := c.MigrateRiver(cmd.Context()); err != nil {
				return fmt.Errorf("run river migrations: %w", err)
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
