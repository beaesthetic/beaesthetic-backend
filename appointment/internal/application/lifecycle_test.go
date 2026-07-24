package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

type lifecycleTxContextKey struct{}

type reminderSchedulerStub struct {
	err                 error
	called              bool
	unscheduleCalled    bool
	txContext           bool
	unscheduleTxContext bool
}

func (s *reminderSchedulerStub) ScheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent, sendAt time.Time) error {
	s.called = true
	s.txContext, _ = ctx.Value(lifecycleTxContextKey{}).(bool)
	return s.err
}

func (s *reminderSchedulerStub) UnscheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent) error {
	s.unscheduleCalled = true
	s.unscheduleTxContext, _ = ctx.Value(lifecycleTxContextKey{}).(bool)
	return nil
}

type notificationSenderStub struct{}

func (s notificationSenderStub) SendAppointmentReminder(context.Context, *domain.AgendaEvent) (string, error) {
	return "", nil
}

func (s notificationSenderStub) SendAppointmentConfirmation(context.Context, *domain.AgendaEvent) (string, error) {
	return "", nil
}

func (s notificationSenderStub) SendAppointmentRescheduled(context.Context, *domain.AgendaEvent) (string, error) {
	return "", nil
}

func TestComputeReminderSendAtReturnsPotentialDateWhenFarEnough(t *testing.T) {
	now := time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)
	eventAt := now.Add(24 * time.Hour)
	sendAt, ok := computeReminderSendAt(now, eventAt, 2*time.Hour, 30*time.Minute, 2*time.Minute)
	if !ok {
		t.Fatal("expected reminder to be schedulable")
	}
	if sendAt == nil {
		t.Fatal("expected sendAt to be set")
	}
	want := eventAt.Add(-2 * time.Hour)
	if !sendAt.Equal(want) {
		t.Fatalf("sendAt=%s want=%s", sendAt, want)
	}
}

func TestComputeReminderSendAtReturnsImmediateDateWhenLateButStillSendable(t *testing.T) {
	now := time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)
	eventAt := now.Add(90 * time.Minute)
	sendAt, ok := computeReminderSendAt(now, eventAt, 2*time.Hour, 30*time.Minute, 2*time.Minute)
	if !ok {
		t.Fatal("expected reminder to be schedulable")
	}
	if sendAt == nil {
		t.Fatal("expected sendAt to be set")
	}
	want := now.Add(2 * time.Minute)
	if !sendAt.Equal(want) {
		t.Fatalf("sendAt=%s want=%s", sendAt, want)
	}
}

func TestComputeReminderSendAtReturnsUnprocessableWhenTooLate(t *testing.T) {
	now := time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)
	eventAt := now.Add(20 * time.Minute)
	sendAt, ok := computeReminderSendAt(now, eventAt, 2*time.Hour, 30*time.Minute, 2*time.Minute)
	if ok {
		t.Fatal("expected reminder to be unprocessable")
	}
	if sendAt != nil {
		t.Fatalf("expected nil sendAt, got %s", sendAt)
	}
}

func TestScheduleReminderSchedulesAndSavesInTransaction(t *testing.T) {
	repo := &appointmentRepoStub{txContextKey: lifecycleTxContextKey{}}
	service := NewAppointmentService(repo, &customerRegistryStub{}, 2*time.Hour, fixedClock{now: time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)})
	scheduler := &reminderSchedulerStub{}
	handler := NewAppointmentLifecycleHandler(
		service,
		repo,
		scheduler,
		notificationSenderStub{},
		fixedClock{now: time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)},
		30*time.Minute,
		2*time.Minute,
		nil,
	)
	event := newLifecycleTestAgendaEvent(t)

	if err := handler.ScheduleReminder(context.Background(), event); err != nil {
		t.Fatalf("ScheduleReminder() error = %v", err)
	}
	if !scheduler.called {
		t.Fatal("expected scheduler to be called")
	}
	if !scheduler.txContext {
		t.Fatal("expected scheduler to be called inside repository transaction")
	}
	if repo.saved == nil {
		t.Fatal("expected agenda event to be saved")
	}
	if repo.saved.ReminderStatus != domain.ReminderScheduled {
		t.Fatalf("reminder status = %s, want %s", repo.saved.ReminderStatus, domain.ReminderScheduled)
	}
	if !repo.saveTxContext {
		t.Fatal("expected agenda event save to happen inside repository transaction")
	}
}

func TestScheduleReminderDoesNotSaveScheduledStateWhenSchedulerFails(t *testing.T) {
	repo := &appointmentRepoStub{txContextKey: lifecycleTxContextKey{}}
	service := NewAppointmentService(repo, &customerRegistryStub{}, 2*time.Hour, fixedClock{now: time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)})
	scheduler := &reminderSchedulerStub{err: errors.New("schedule failed")}
	handler := NewAppointmentLifecycleHandler(
		service,
		repo,
		scheduler,
		notificationSenderStub{},
		fixedClock{now: time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)},
		30*time.Minute,
		2*time.Minute,
		nil,
	)
	event := newLifecycleTestAgendaEvent(t)

	err := handler.ScheduleReminder(context.Background(), event)
	if err == nil {
		t.Fatal("ScheduleReminder() error is nil")
	}
	if repo.saved != nil {
		t.Fatalf("expected no saved scheduled state, got %#v", repo.saved)
	}
}

func TestHandleScheduledRunsReminderAndConfirmationInSingleTransaction(t *testing.T) {
	repo := &appointmentRepoStub{txContextKey: lifecycleTxContextKey{}}
	service := NewAppointmentService(repo, &customerRegistryStub{}, 2*time.Hour, fixedClock{now: time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)})
	scheduler := &reminderSchedulerStub{}
	handler := NewAppointmentLifecycleHandler(
		service,
		repo,
		scheduler,
		notificationSenderStub{},
		fixedClock{now: time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)},
		30*time.Minute,
		2*time.Minute,
		nil,
	)
	event := newLifecycleTestAgendaEvent(t)

	if err := handler.handleScheduled(context.Background(), event, false); err != nil {
		t.Fatalf("handleScheduled() error = %v", err)
	}
	if repo.txCalls != 1 {
		t.Fatalf("tx calls = %d, want 1", repo.txCalls)
	}
	if !scheduler.txContext {
		t.Fatal("expected scheduler to be called inside repository transaction")
	}
	if !repo.saveTxContext {
		t.Fatal("expected agenda event save to happen inside repository transaction")
	}
	if repo.pendingNotification == nil {
		t.Fatal("expected pending confirmation notification to be tracked")
	}
	if !repo.pendingTxContext {
		t.Fatal("expected pending notification tracking to happen inside repository transaction")
	}
}

func TestHandleDeletedUnschedulesReminderInTransaction(t *testing.T) {
	repo := &appointmentRepoStub{txContextKey: lifecycleTxContextKey{}}
	service := NewAppointmentService(repo, &customerRegistryStub{}, 2*time.Hour, fixedClock{now: time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)})
	scheduler := &reminderSchedulerStub{}
	handler := NewAppointmentLifecycleHandler(
		service,
		repo,
		scheduler,
		notificationSenderStub{},
		fixedClock{now: time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)},
		30*time.Minute,
		2*time.Minute,
		nil,
	)
	event := newLifecycleTestAgendaEvent(t)

	if err := handler.handleDeleted(context.Background(), event); err != nil {
		t.Fatalf("handleDeleted() error = %v", err)
	}
	if !scheduler.unscheduleCalled {
		t.Fatal("expected scheduler unschedule to be called")
	}
	if !scheduler.unscheduleTxContext {
		t.Fatal("expected scheduler unschedule to be called inside repository transaction")
	}
	if repo.saved == nil {
		t.Fatal("expected agenda event to be saved")
	}
	if repo.saved.ReminderStatus != domain.ReminderDeleted {
		t.Fatalf("reminder status = %s, want %s", repo.saved.ReminderStatus, domain.ReminderDeleted)
	}
	if !repo.saveTxContext {
		t.Fatal("expected agenda event save to happen inside repository transaction")
	}
}

func newLifecycleTestAgendaEvent(t *testing.T) *domain.AgendaEvent {
	t.Helper()
	event, err := domain.NewAgendaEvent(
		"event-1",
		domain.EventTypeAppointment,
		"Haircut",
		"",
		time.Date(2026, time.July, 5, 12, 0, 0, 0, time.UTC),
		time.Date(2026, time.July, 5, 13, 0, 0, 0, time.UTC),
		domain.Attendee{ID: "customer-1", DisplayName: "Jane Doe"},
		nil,
		2*time.Hour,
		time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	return &event
}
