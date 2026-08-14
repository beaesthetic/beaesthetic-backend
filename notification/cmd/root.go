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
	appruntime "github.com/petretiandrea/beaesthetic-backend/core-contracts/runtime"
	"github.com/petretiandrea/beaesthetic-backend/notification/cmd/di"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	var envFile string
	root := &cobra.Command{
		Use:          "notification",
		Short:        "Notification service",
		SilenceUsage: true,
	}
	root.PersistentFlags().StringVar(&envFile, "env-file", "", "optional dotenv file")
	root.AddCommand(appCommand(&envFile), migrateCommand(&envFile))
	return root
}

func appCommand(envFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "app",
		Short: "Start the HTTP API and RabbitMQ consumer",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			var customerNotificationConsumer interface {
				Run(context.Context) error
			}
			if c.Config.RabbitMQ.CustomerNotificationQueue != "" {
				customerNotificationConsumer = c.GetCustomerNotificationConsumer()
			}

			runner := appruntime.NewRunner(c.Log)
			runner.Add(appruntime.HTTPServer("http server", httpServer, 10*time.Second))
			if customerNotificationConsumer != nil {
				runner.Add(appruntime.Consumer("customer notifications consumer", customerNotificationConsumer))
			}

			return runner.Run(ctx)
		},
	}
}

func migrateCommand(envFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate [up|down|version]",
		Short: "Run PostgreSQL migrations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
