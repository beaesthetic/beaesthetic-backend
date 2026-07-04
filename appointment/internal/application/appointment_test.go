package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

type appointmentRepoStub struct {
	saved *domain.AgendaEvent
}

func (r *appointmentRepoStub) Tx(ctx context.Context, atomicFn func(context.Context) error) error {
	return atomicFn(ctx)
}

func (r *appointmentRepoStub) SaveAgendaEvent(ctx context.Context, event *domain.AgendaEvent) error {
	r.saved = event
	return nil
}

func (r *appointmentRepoStub) FindAgendaEvent(context.Context, string) (*domain.AgendaEvent, error) {
	return nil, nil
}

func (r *appointmentRepoStub) SearchAgendaEvents(context.Context, string, *time.Time, *time.Time) ([]domain.AgendaEvent, error) {
	return nil, nil
}

func (r *appointmentRepoStub) FindPendingNotification(context.Context, string) (*PendingNotification, error) {
	return nil, nil
}

func (r *appointmentRepoStub) RemovePendingNotification(context.Context, string) error {
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
