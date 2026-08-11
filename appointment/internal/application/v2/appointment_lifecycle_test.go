package v2

import (
	"context"
	"testing"
	"time"

	domain "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
)

type calendarReminderSchedulerStub struct {
	scheduled       bool
	unscheduled     bool
	eventID         string
	expectedStartAt time.Time
	sendAt          time.Time
}

func (s *calendarReminderSchedulerStub) ScheduleCalendarReminder(_ context.Context, eventID string, expectedStartAt time.Time, sendAt time.Time) error {
	s.scheduled = true
	s.eventID = eventID
	s.expectedStartAt = expectedStartAt
	s.sendAt = sendAt
	return nil
}

func (s *calendarReminderSchedulerStub) UnscheduleCalendarReminder(_ context.Context, eventID string) error {
	s.unscheduled = true
	s.eventID = eventID
	return nil
}

type calendarNotificationSenderStub struct {
	calls            int
	notificationType domain.NotificationType
	idempotencyKey   string
}

func (s *calendarNotificationSenderStub) SendCalendarNotification(_ context.Context, _ domain.CalendarEvent, notificationType domain.NotificationType, idempotencyKey string) (string, error) {
	s.calls++
	s.notificationType = notificationType
	s.idempotencyKey = idempotencyKey
	return idempotencyKey, nil
}

func TestAppointmentLifecycleCreatedSchedulesAndTracksConfirmationAtomically(t *testing.T) {
	now := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	event := newAppointmentLifecycleEvent(t, now.Add(48*time.Hour), now.Add(49*time.Hour), now)
	repository := &repositoryStub{
		found:     &event,
		reminders: map[string]domain.AppointmentReminder{event.ID: mustReminder(t, 24*time.Hour, now)},
	}
	scheduler := &calendarReminderSchedulerStub{}
	notifications := &calendarNotificationSenderStub{}
	service := NewAppointmentLifecycleService(repository, scheduler, notifications, clockStub{now: now}, 30*time.Minute, 2*time.Minute)

	if err := service.Handle(context.Background(), "CalendarEventCreated", event.ID); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	reminder := repository.reminders[event.ID]
	if !scheduler.scheduled || reminder.Status != domain.ReminderStatusScheduled || reminder.ScheduledAt == nil {
		t.Fatalf("scheduler=%v reminder=%#v", scheduler.scheduled, reminder)
	}
	if notifications.notificationType != domain.NotificationTypeAppointmentConfirmation || len(repository.notifications) != 1 {
		t.Fatalf("notification type=%s tracked=%d", notifications.notificationType, len(repository.notifications))
	}
	if repository.txCalls != 1 || repository.writesOutsideTx != 0 {
		t.Fatalf("tx=%d writes outside=%d", repository.txCalls, repository.writesOutsideTx)
	}
}

func TestAppointmentLifecycleIgnoresManualEvents(t *testing.T) {
	now := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	eventRange, err := domain.NewTimeRange(now.Add(time.Hour), now.Add(2*time.Hour), "Europe/Rome", false)
	if err != nil {
		t.Fatal(err)
	}
	event, err := domain.NewManualCalendarEvent(domain.ManualEventParams{EventID: "event-1", CalendarID: domain.DefaultCalendarID, Range: eventRange, Title: "Work", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	repository := &repositoryStub{found: &event}
	scheduler := &calendarReminderSchedulerStub{}
	notifications := &calendarNotificationSenderStub{}
	service := NewAppointmentLifecycleService(repository, scheduler, notifications, clockStub{now: now}, 30*time.Minute, 2*time.Minute)

	if err := service.Handle(context.Background(), "CalendarEventCreated", event.ID); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if scheduler.scheduled || notifications.calls != 0 {
		t.Fatalf("manual event triggered appointment lifecycle")
	}
}

func TestAppointmentLifecycleOutcomeUpdatesNotificationAndReminder(t *testing.T) {
	now := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	event := newAppointmentLifecycleEvent(t, now.Add(48*time.Hour), now.Add(49*time.Hour), now)
	appointment := event.Detail.(domain.Appointment)
	recipient, err := domain.NewCustomerNotificationRecipient(appointment.Customer.ID)
	if err != nil {
		t.Fatal(err)
	}
	notification, err := domain.NewAppointmentNotification("notification-1", event.ID, domain.NotificationKindReminder, recipient, nil, now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	repository := &repositoryStub{
		found:         &event,
		reminders:     map[string]domain.AppointmentReminder{event.ID: mustReminder(t, 24*time.Hour, now)},
		notifications: map[string]domain.AppointmentNotification{notification.CorrelationKey: notification},
	}
	service := NewAppointmentLifecycleService(repository, &calendarReminderSchedulerStub{}, &calendarNotificationSenderStub{}, clockStub{now: now.Add(time.Hour)}, 30*time.Minute, 2*time.Minute)

	eventID, err := service.HandleNotificationOutcome(context.Background(), notification.CorrelationKey, true, "", "")
	if err != nil {
		t.Fatalf("HandleNotificationOutcome() error = %v", err)
	}
	if eventID != event.ID || repository.notifications[notification.CorrelationKey].Status != domain.NotificationStatusSent || repository.reminders[event.ID].Status != domain.ReminderStatusSent {
		t.Fatalf("event=%s notification=%#v reminder=%#v", eventID, repository.notifications[notification.CorrelationKey], repository.reminders[event.ID])
	}
}

func TestAppointmentLifecycleSendsDueReminderWithoutLegacyDomain(t *testing.T) {
	now := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	event := newAppointmentLifecycleEvent(t, now.Add(24*time.Hour), now.Add(25*time.Hour), now)
	reminder := mustReminder(t, 2*time.Hour, now)
	if err := reminder.Schedule(now.Add(22*time.Hour), now); err != nil {
		t.Fatal(err)
	}
	repository := &repositoryStub{found: &event, reminders: map[string]domain.AppointmentReminder{event.ID: reminder}}
	notifications := &calendarNotificationSenderStub{}
	service := NewAppointmentLifecycleService(repository, &calendarReminderSchedulerStub{}, notifications, clockStub{now: now.Add(22 * time.Hour)}, 30*time.Minute, 2*time.Minute)

	if err := service.SendDueReminder(context.Background(), event.ID, &event.Range.Start); err != nil {
		t.Fatalf("SendDueReminder() error = %v", err)
	}
	if notifications.notificationType != domain.NotificationTypeAppointmentReminder || repository.reminders[event.ID].Status != domain.ReminderStatusSendRequested || len(repository.notifications) != 1 {
		t.Fatalf("notification=%s reminder=%#v tracked=%d", notifications.notificationType, repository.reminders[event.ID], len(repository.notifications))
	}
}

func TestAppointmentLifecycleResendUsesRequestIdempotencyKey(t *testing.T) {
	now := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	event := newAppointmentLifecycleEvent(t, now.Add(24*time.Hour), now.Add(25*time.Hour), now)
	repository := &repositoryStub{found: &event, reminders: map[string]domain.AppointmentReminder{event.ID: mustReminder(t, 2*time.Hour, now)}}
	notifications := &calendarNotificationSenderStub{}
	service := NewAppointmentLifecycleService(repository, &calendarReminderSchedulerStub{}, notifications, clockStub{now: now}, 30*time.Minute, 2*time.Minute)

	if err := service.RequestReminderResend(context.Background(), event.ID, "request-1"); err != nil {
		t.Fatalf("RequestReminderResend() error = %v", err)
	}
	wantKey := "appointment:event-1:reminder:resend:request-1"
	if notifications.idempotencyKey != wantKey || repository.reminders[event.ID].Status != domain.ReminderStatusSendRequested {
		t.Fatalf("idempotency=%q reminder=%#v", notifications.idempotencyKey, repository.reminders[event.ID])
	}
	if notification, ok := repository.notifications[wantKey]; !ok || notification.IdempotencyKey == nil || *notification.IdempotencyKey != wantKey {
		t.Fatalf("tracked notification=%#v", notification)
	}
}

func newAppointmentLifecycleEvent(t *testing.T, start time.Time, end time.Time, now time.Time) domain.CalendarEvent {
	t.Helper()
	eventRange, err := domain.NewTimeRange(start, end, "Europe/Rome", false)
	if err != nil {
		t.Fatal(err)
	}
	customer, err := domain.NewCustomerRef("customer-1", "Jane Doe")
	if err != nil {
		t.Fatal(err)
	}
	event, err := domain.NewAppointmentEvent(domain.AppointmentEventParams{EventID: "event-1", CalendarID: domain.DefaultCalendarID, Range: eventRange, Customer: customer, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	event.PullEvents()
	return event
}

func mustReminder(t *testing.T, remindBefore time.Duration, now time.Time) domain.AppointmentReminder {
	t.Helper()
	reminder, err := domain.NewAppointmentReminder(remindBefore, now)
	if err != nil {
		t.Fatal(err)
	}
	return reminder
}
