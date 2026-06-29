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
	"github.com/beaesthetic/scheduler/internal/infra/redislegacy"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main() {
	var configPath string
	var migrationsPath string
	var dryRun bool

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

	migrateOldCmd := &cobra.Command{
		Use:   "migrate-old",
		Short: "copy legacy Redis schedules into Postgres",
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

			ctx := cmd.Context()
			pool, err := postgres.Open(ctx, cfg.Postgres.DSN)
			if err != nil {
				return err
			}
			defer pool.Close()

			store := postgres.NewJobRepository(pool, cfg.Scheduler.Name)
			migrator := redislegacy.NewMigrator(cfg, store, logger)
			result, err := migrator.Run(ctx, dryRun)
			if err != nil {
				return err
			}
			logger.Info("legacy Redis migration completed",
				zap.Int("copied", result.Copied),
				zap.Int("skipped", result.Skipped),
				zap.Int("failed", result.Failed),
				zap.Bool("dry_run", dryRun),
			)
			return nil
		},
	}
	migrateOldCmd.Flags().BoolVar(&dryRun, "dry-run", false, "read legacy Redis data without writing Postgres")
	rootCmd.AddCommand(migrateOldCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
