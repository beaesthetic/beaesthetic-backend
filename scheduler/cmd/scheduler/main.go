package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/beaesthetic/scheduler/internal/config"
	"github.com/beaesthetic/scheduler/internal/container"
	"github.com/beaesthetic/scheduler/internal/infra/postgres"
	"github.com/spf13/cobra"
)

func main() {
	var configPath string
	var migrationsPath string

	rootCmd := &cobra.Command{
		Use:   "scheduler",
		Short: "Scheduler service",
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "optional YAML config file")
	rootCmd.PersistentFlags().StringVar(&migrationsPath, "migrations", "file://migrations", "golang-migrate migrations source URL")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "app",
		Short: "start scheduler service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			c, err := container.New(ctx, cfg)
			if err != nil {
				return err
			}
			defer func() {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				_ = c.Close(shutdownCtx)
			}()

			return c.Run(ctx)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "migrate",
		Short: "run Postgres schema migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			logger, err := container.NewLogger(cfg)
			if err != nil {
				return err
			}
			defer logger.Sync()

			logger.Info("running scheduler migrations")
			if err := postgres.RunMigrations(cfg.Postgres.DSN, migrationsPath); err != nil {
				return err
			}
			logger.Info("scheduler migrations completed")
			return nil
		},
	})

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
