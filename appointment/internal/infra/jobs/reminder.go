package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

	reminders DueReminderSender
}

type DueReminderSender interface {
	SendDueReminder(ctx context.Context, calendarEventID string, expectedStartAt *time.Time) error
}

func NewSendAppointmentReminderWorker(reminders DueReminderSender) *SendAppointmentReminderWorker {
	return &SendAppointmentReminderWorker{reminders: reminders}
}

func (w *SendAppointmentReminderWorker) Work(ctx context.Context, job *river.Job[SendAppointmentReminderArgs]) error {
	return w.reminders.SendDueReminder(ctx, job.Args.EventID, &job.Args.ExpectedStartAt)
}

type JobInserter interface {
	Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) error
	CancelByKey(ctx context.Context, kind string, queue string, key string) error
}

type ReminderScheduler struct {
	inserter    JobInserter
	queue       string
	maxAttempts int
	log         *zap.Logger
}

func NewReminderScheduler(inserter JobInserter, queue string, maxAttempts int, log *zap.Logger) *ReminderScheduler {
	if log == nil {
		log = zap.NewNop()
	}
	return &ReminderScheduler{
		inserter:    inserter,
		queue:       queue,
		maxAttempts: maxAttempts,
		log:         log.Named("river_reminder_scheduler"),
	}
}

func (s *ReminderScheduler) ScheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent, sendAt time.Time) error {
	return s.ScheduleCalendarReminder(ctx, agendaEvent.ID, agendaEvent.Start, sendAt)
}

func (s *ReminderScheduler) ScheduleCalendarReminder(ctx context.Context, calendarEventID string, expectedStartAt time.Time, sendAt time.Time) error {
	key := appointmentReminderKey(calendarEventID)
	if err := s.inserter.CancelByKey(ctx, SendAppointmentReminderKind, s.queue, key); err != nil {
		return err
	}
	return s.inserter.Insert(ctx, SendAppointmentReminderArgs{
		EventID:         calendarEventID,
		ExpectedStartAt: expectedStartAt.UTC(),
	}, &river.InsertOpts{
		Queue:       s.queue,
		ScheduledAt: sendAt.UTC(),
		MaxAttempts: s.maxAttempts,
		Metadata:    appointmentReminderMetadata(key),
	})
}

func (s *ReminderScheduler) UnscheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent) error {
	return s.UnscheduleCalendarReminder(ctx, agendaEvent.ID)
}

func (s *ReminderScheduler) UnscheduleCalendarReminder(ctx context.Context, calendarEventID string) error {
	return s.inserter.CancelByKey(ctx, SendAppointmentReminderKind, s.queue, appointmentReminderKey(calendarEventID))
}

func appointmentReminderKey(eventID string) string {
	return fmt.Sprintf("appointment:%s:reminder", eventID)
}

func appointmentReminderMetadata(key string) []byte {
	metadata, _ := json.Marshal(map[string]string{"idempotencyKey": key})
	return metadata
}
