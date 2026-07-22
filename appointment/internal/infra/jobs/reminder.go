package jobs

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

const SendAppointmentReminderKind = "appointment.send_reminder"

type SendAppointmentReminderArgs struct {
	EventID         string    `json:"eventId"`
	ExpectedStartAt time.Time `json:"expectedStartAt"`
}

func (SendAppointmentReminderArgs) Kind() string {
	return SendAppointmentReminderKind
}

type SendAppointmentReminderWorker struct {
	river.WorkerDefaults[SendAppointmentReminderArgs]

	reminders *application.ReminderSender
}

func NewSendAppointmentReminderWorker(reminders *application.ReminderSender) *SendAppointmentReminderWorker {
	return &SendAppointmentReminderWorker{reminders: reminders}
}

func (w *SendAppointmentReminderWorker) Work(ctx context.Context, job *river.Job[SendAppointmentReminderArgs]) error {
	return w.reminders.SendDueReminder(ctx, job.Args.EventID, &job.Args.ExpectedStartAt)
}

type ReminderScheduler struct {
	client      *river.Client[pgx.Tx]
	queue       string
	maxAttempts int
	log         *zap.Logger
}

func NewReminderScheduler(client *river.Client[pgx.Tx], queue string, maxAttempts int, log *zap.Logger) *ReminderScheduler {
	if log == nil {
		log = zap.NewNop()
	}
	return &ReminderScheduler{
		client:      client,
		queue:       queue,
		maxAttempts: maxAttempts,
		log:         log.Named("river_reminder_scheduler"),
	}
}

func (s *ReminderScheduler) ScheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent, sendAt time.Time) error {
	_, err := s.client.Insert(ctx, SendAppointmentReminderArgs{
		EventID:         agendaEvent.ID,
		ExpectedStartAt: agendaEvent.Start.UTC(),
	}, &river.InsertOpts{
		Queue:       s.queue,
		ScheduledAt: sendAt.UTC(),
		MaxAttempts: s.maxAttempts,
	})
	return err
}

func (s *ReminderScheduler) UnscheduleReminder(ctx context.Context, eventID string) error {
	s.log.Debug("river reminder unschedule requested; stale jobs are skipped by worker validation", zap.String("event_id", eventID))
	return nil
}
