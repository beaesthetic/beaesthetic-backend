package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/petretiandrea/beaesthetic-backend/appointment/cmd/di"
	appruntime "github.com/petretiandrea/beaesthetic-backend/core-contracts/runtime"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewRootCommand() *cobra.Command {
	var envFile string
	root := &cobra.Command{Use: "appointment", Short: "Appointment service", SilenceUsage: true}
	root.PersistentFlags().StringVar(&envFile, "env-file", "", "optional dotenv file")
	root.AddCommand(appCommand(&envFile), migrateCommand(&envFile), scheduleFutureRemindersCommand(&envFile))
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
		appointmentLifecycleConsumer := c.GetAppointmentLifecycleConsumer()
		schedulerConsumer := c.GetSchedulerQueueConsumer()
		notificationOutcomeConsumer := c.GetNotificationOutcomeQueueConsumer()

		runner := appruntime.NewRunner(c.Log)
		riverClient := c.GetRiverClient()
		runner.Add(appruntime.StartStop("river reminder scheduler", riverClient.Start, riverClient.Stop, 10*time.Second))
		runner.Add(appruntime.HTTPServer("http server", httpServer, 10*time.Second))
		runner.Add(appruntime.Consumer("appointment lifecycle consumer", appointmentLifecycleConsumer))
		runner.Add(appruntime.Consumer("scheduler queue consumer", schedulerConsumer))
		runner.Add(appruntime.Consumer("notification outcome consumer", notificationOutcomeConsumer))

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

func scheduleFutureRemindersCommand(envFile *string) *cobra.Command {
	var dryRun bool
	var from string
	cmd := &cobra.Command{Use: "schedule-future-reminders", Short: "Schedule reminders for future appointments after a manual backfill", RunE: func(cmd *cobra.Command, args []string) error {
		c, err := di.NewDiContainer(cmd.Context(), *envFile)
		if err != nil {
			return err
		}
		defer func() {
			c.GetPostgresDatabase().Close()
			_ = c.Log.Sync()
		}()

		fromTime := c.GetClock().Now().UTC()
		if from != "" {
			parsedFrom, err := parseScheduleFutureRemindersFrom(from)
			if err != nil {
				return err
			}
			fromTime = parsedFrom
		}

		appointments, err := c.GetAppointmentService().FutureAppointments(cmd.Context(), fromTime)
		if err != nil {
			return fmt.Errorf("find future appointments: %w", err)
		}

		log := c.Log.Named("manual_schedule_future_reminders")
		log.Info("found future appointments", zap.Int("count", len(appointments)), zap.Time("from", fromTime), zap.Bool("dry_run", dryRun))

		var scheduled int
		var failed int
		for i := range appointments {
			appointment := &appointments[i]
			if dryRun {
				log.Info("would schedule appointment reminder", zap.String("event_id", appointment.ID), zap.Time("start_at", appointment.Start), zap.Duration("remind_before", appointment.RemindBefore), zap.String("reminder_status", string(appointment.ReminderStatus)))
				continue
			}
			if err := c.GetAppointmentLifecycleHandler().ScheduleReminder(cmd.Context(), appointment); err != nil {
				failed++
				log.Error("failed to schedule appointment reminder", zap.String("event_id", appointment.ID), zap.Time("start_at", appointment.Start), zap.Error(err))
				continue
			}
			scheduled++
		}

		log.Info("manual future reminder scheduling completed", zap.Int("found", len(appointments)), zap.Int("scheduled", scheduled), zap.Int("failed", failed), zap.Bool("dry_run", dryRun))
		if failed > 0 {
			return fmt.Errorf("failed to schedule %d future reminders", failed)
		}
		return nil
	}}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "log future appointments without scheduling reminders")
	cmd.Flags().StringVar(&from, "from", "", "schedule only appointments starting from this date/time (YYYY-MM-DD or RFC3339)")
	return cmd
}

func parseScheduleFutureRemindersFrom(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), nil
	}
	parsedDate, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid --from %q: use YYYY-MM-DD or RFC3339", value)
	}
	return parsedDate.UTC(), nil
}
