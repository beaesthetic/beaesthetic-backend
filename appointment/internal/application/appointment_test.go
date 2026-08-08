package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

type appointmentRepoStub struct {
	saved                 *domain.AgendaEvent
	savedReminder         *domain.AgendaEvent
	txContextKey          any
	txCalls               int
	saveTxContext         bool
	saveReminderTxContext bool
	pendingTxContext      bool
	pendingNotification   *AppointmentNotificationTracking
	pendingLookup         *AppointmentNotificationTracking
	agendaLookup          *domain.AgendaEvent
	sentPending           string
	failedPending         string
}

func (r *appointmentRepoStub) Tx(ctx context.Context, atomicFn func(context.Context) error) error {
	r.txCalls++
	if r.txContextKey != nil {
		ctx = context.WithValue(ctx, r.txContextKey, true)
	}
	return atomicFn(ctx)
}

func (r *appointmentRepoStub) SaveAgendaEvent(ctx context.Context, event *domain.AgendaEvent) error {
	r.saved = event
	if r.txContextKey != nil {
		r.saveTxContext, _ = ctx.Value(r.txContextKey).(bool)
	}
	return nil
}

func (r *appointmentRepoStub) SaveAppointmentReminder(ctx context.Context, event *domain.AgendaEvent) error {
	r.savedReminder = event
	if r.txContextKey != nil {
		r.saveReminderTxContext, _ = ctx.Value(r.txContextKey).(bool)
	}
	return nil
}

func (r *appointmentRepoStub) FindAgendaEvent(context.Context, string) (*domain.AgendaEvent, error) {
	return r.agendaLookup, nil
}

func (r *appointmentRepoStub) SearchAgendaEvents(context.Context, string, *time.Time, *time.Time) ([]domain.AgendaEvent, error) {
	return nil, nil
}

func (r *appointmentRepoStub) FindFutureAppointments(context.Context, time.Time) ([]domain.AgendaEvent, error) {
	return nil, nil
}
func (r *appointmentRepoStub) FindAppointmentNotificationTracking(context.Context, string) (*AppointmentNotificationTracking, error) {
	return r.pendingLookup, nil
}

func (r *appointmentRepoStub) SaveAppointmentNotificationTracking(ctx context.Context, pending AppointmentNotificationTracking) error {
	r.pendingNotification = &pending
	if r.txContextKey != nil {
		r.pendingTxContext, _ = ctx.Value(r.txContextKey).(bool)
	}
	return nil
}

func (r *appointmentRepoStub) MarkAppointmentNotificationSent(ctx context.Context, correlationKey string, completedAt time.Time) error {
	r.sentPending = correlationKey
	return nil
}

func (r *appointmentRepoStub) MarkAppointmentNotificationFailed(ctx context.Context, correlationKey string, reason string, message string, completedAt time.Time) error {
	r.failedPending = correlationKey
	return nil
}

type customerRegistryStub struct {
	customer *Customer
	calls    int
}

func (r *customerRegistryStub) FindByCustomerID(context.Context, string) (*Customer, error) {
	r.calls++
	return r.customer, nil
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

func TestCreateAgendaAppointmentResolvesAttendeeFromCustomerRegistry(t *testing.T) {
	repo := &appointmentRepoStub{}
	registry := &customerRegistryStub{customer: &Customer{ID: "customer-1", DisplayName: "Jane Doe"}}
	service := NewAppointmentService(repo, registry, time.Hour, fixedClock{now: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)})

	_, err := service.CreateAgenda(
		context.Background(),
		domain.EventTypeAppointment,
		"Haircut",
		"",
		time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 5, 11, 0, 0, 0, time.UTC),
		domain.Attendee{ID: "customer-1"},
		nil,
	)
	if err != nil {
		t.Fatalf("CreateAgenda returned error: %v", err)
	}
	if registry.calls != 1 {
		t.Fatalf("expected customer registry to be called once, got %d", registry.calls)
	}
	if repo.saved == nil || repo.saved.Attendee.DisplayName != "Jane Doe" {
		t.Fatalf("expected saved attendee display name from customer, got %#v", repo.saved)
	}
}

func TestCreateAgendaGenericEventDoesNotResolveCustomer(t *testing.T) {
	repo := &appointmentRepoStub{}
	registry := &customerRegistryStub{customer: &Customer{ID: "customer-1", DisplayName: "Jane Doe"}}
	service := NewAppointmentService(repo, registry, time.Hour, fixedClock{now: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)})

	_, err := service.CreateAgenda(
		context.Background(),
		domain.EventTypeGeneric,
		"Blocked slot",
		"",
		time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 5, 11, 0, 0, 0, time.UTC),
		domain.Attendee{ID: "self-id"},
		nil,
	)
	if err != nil {
		t.Fatalf("CreateAgenda returned error: %v", err)
	}
	if registry.calls != 0 {
		t.Fatalf("expected customer registry not to be called, got %d calls", registry.calls)
	}
	if repo.saved == nil || repo.saved.Attendee.DisplayName != "self" {
		t.Fatalf("expected generic attendee display name self, got %#v", repo.saved)
	}
}

func TestCreateAgendaAppointmentFailsWhenCustomerIsUnknown(t *testing.T) {
	repo := &appointmentRepoStub{}
	registry := &customerRegistryStub{}
	service := NewAppointmentService(repo, registry, time.Hour, fixedClock{now: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)})

	_, err := service.CreateAgenda(
		context.Background(),
		domain.EventTypeAppointment,
		"Haircut",
		"",
		time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 5, 11, 0, 0, 0, time.UTC),
		domain.Attendee{ID: "missing-customer"},
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "Unknown attendee missing-customer") {
		t.Fatalf("expected unknown attendee error, got %v", err)
	}
	if repo.saved != nil {
		t.Fatalf("expected no saved event, got %#v", repo.saved)
	}
}

func TestConfirmNotificationMarksReminderSent(t *testing.T) {
	event := newAppointmentTestAgendaEvent(t)
	event.ReminderStatus = domain.ReminderSentRequested
	repo := &appointmentRepoStub{
		pendingLookup: &AppointmentNotificationTracking{CorrelationKey: "notification-1", AgendaEventID: event.ID, Kind: "reminder", Type: NotificationTypeAppointmentReminder},
		agendaLookup:  event,
	}
	service := NewAppointmentService(repo, &customerRegistryStub{}, time.Hour, fixedClock{now: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)})

	agendaEvent, err := service.ConfirmNotification(context.Background(), "notification-1")
	if err != nil {
		t.Fatalf("ConfirmNotification() error = %v", err)
	}
	if agendaEvent == nil || agendaEvent.ReminderStatus != domain.ReminderSent {
		t.Fatalf("reminder status = %v, want %s", agendaEvent, domain.ReminderSent)
	}
	if repo.sentPending != "notification-1" {
		t.Fatalf("sent pending = %q, want notification-1", repo.sentPending)
	}
}

func TestFailNotificationMarksReminderFailToSend(t *testing.T) {
	event := newAppointmentTestAgendaEvent(t)
	event.ReminderStatus = domain.ReminderSentRequested
	repo := &appointmentRepoStub{
		pendingLookup: &AppointmentNotificationTracking{CorrelationKey: "notification-1", AgendaEventID: event.ID, Kind: "reminder", Type: NotificationTypeAppointmentReminder},
		agendaLookup:  event,
	}
	service := NewAppointmentService(repo, &customerRegistryStub{}, time.Hour, fixedClock{now: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)})

	agendaEvent, err := service.FailNotification(context.Background(), "notification-1")
	if err != nil {
		t.Fatalf("FailNotification() error = %v", err)
	}
	if agendaEvent == nil || agendaEvent.ReminderStatus != domain.ReminderFailToSend {
		t.Fatalf("reminder status = %v, want %s", agendaEvent, domain.ReminderFailToSend)
	}
	if repo.failedPending != "notification-1" {
		t.Fatalf("failed pending = %q, want notification-1", repo.failedPending)
	}
}

func TestConfirmNotificationDoesNotUpdateReminderForConfirmation(t *testing.T) {
	event := newAppointmentTestAgendaEvent(t)
	event.ReminderStatus = domain.ReminderScheduled
	repo := &appointmentRepoStub{
		pendingLookup: &AppointmentNotificationTracking{CorrelationKey: "notification-1", AgendaEventID: event.ID, Kind: "confirmation", Type: NotificationTypeAppointmentConfirmation},
		agendaLookup:  event,
	}
	service := NewAppointmentService(repo, &customerRegistryStub{}, time.Hour, fixedClock{now: time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC)})

	agendaEvent, err := service.ConfirmNotification(context.Background(), "notification-1")
	if err != nil {
		t.Fatalf("ConfirmNotification() error = %v", err)
	}
	if agendaEvent != nil {
		t.Fatalf("expected no reminder agenda event update, got %#v", agendaEvent)
	}
	if repo.saved != nil {
		t.Fatalf("expected no saved reminder update, got %#v", repo.saved)
	}
	if repo.sentPending != "notification-1" {
		t.Fatalf("sent pending = %q, want notification-1", repo.sentPending)
	}
}

func newAppointmentTestAgendaEvent(t *testing.T) *domain.AgendaEvent {
	t.Helper()
	event, err := domain.NewAgendaEvent(
		"event-1",
		domain.EventTypeAppointment,
		"Haircut",
		"",
		time.Date(2026, 7, 5, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 5, 11, 0, 0, 0, time.UTC),
		domain.Attendee{ID: "customer-1", DisplayName: "Jane Doe"},
		nil,
		time.Hour,
		time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	return &event
}
