package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

type AppointmentRepository interface {
	Tx(ctx context.Context, atomicFn func(ctx context.Context) error) error
	SaveAgendaEvent(context.Context, *domain.AgendaEvent) error
	FindAgendaEvent(context.Context, string) (*domain.AgendaEvent, error)
	SearchAgendaEvents(context.Context, string, *time.Time, *time.Time) ([]domain.AgendaEvent, error)
	FindFutureAppointments(context.Context, time.Time) ([]domain.AgendaEvent, error)
	FindPendingNotification(context.Context, string) (*PendingNotification, error)
	RemovePendingNotification(context.Context, string) error
	SavePendingNotification(context.Context, PendingNotification) error
}

type PendingNotification struct {
	CorrelationKey string
	AgendaEventID  string
	Type           string
}

type AppointmentService struct {
	repo             AppointmentRepository
	customerRegistry CustomerRegistry
	clock            Clock
	reminderBefore   time.Duration
}

func NewAppointmentService(repo AppointmentRepository, customerRegistry CustomerRegistry, reminderBefore time.Duration, clock Clock) *AppointmentService {
	return &AppointmentService{repo: repo, customerRegistry: customerRegistry, clock: clock, reminderBefore: reminderBefore}
}

func (s *AppointmentService) CreateAgenda(ctx context.Context, typ domain.EventType, title, description string, start, end time.Time, attendee domain.Attendee, services []domain.AppointmentServiceRef) (*domain.AgendaEvent, error) {
	resolvedAttendee, err := s.resolveAttendee(ctx, typ, attendee)
	if err != nil {
		return nil, err
	}
	e, err := domain.NewAgendaEvent(uuid.NewString(), typ, title, description, start, end, resolvedAttendee, services, s.reminderBefore, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.saveAgenda(ctx, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *AppointmentService) GetAgenda(ctx context.Context, id string) (*domain.AgendaEvent, error) {
	return s.repo.FindAgendaEvent(ctx, id)
}

func (s *AppointmentService) SaveAgendaEvent(ctx context.Context, event *domain.AgendaEvent) error {
	return s.repo.SaveAgendaEvent(ctx, event)
}

func (s *AppointmentService) SearchAgenda(ctx context.Context, attendee string, start, end *time.Time) ([]domain.AgendaEvent, error) {
	return s.repo.SearchAgendaEvents(ctx, attendee, start, end)
}

func (s *AppointmentService) FutureAppointments(ctx context.Context, from time.Time) ([]domain.AgendaEvent, error) {
	return s.repo.FindFutureAppointments(ctx, from)
}

func (s *AppointmentService) UpdateAgenda(ctx context.Context, id string, title, description *string, start, end *time.Time, services []domain.AppointmentServiceRef) (*domain.AgendaEvent, error) {
	var event *domain.AgendaEvent
	err := s.repo.Tx(ctx, func(ctx context.Context) error {
		e, err := s.repo.FindAgendaEvent(ctx, id)
		if err != nil || e == nil {
			event = e
			return err
		}
		if title != nil {
			e.Title = *title
		}
		if description != nil {
			e.Description = *description
		}
		if start != nil && end != nil {
			if err := e.Reschedule(*start, *end); err != nil {
				return err
			}
		}
		if services != nil {
			e.Services = services
		}
		event = e
		return s.repo.SaveAgendaEvent(ctx, e)
	})
	return event, err
}

func (s *AppointmentService) DeleteAgenda(ctx context.Context, id string, reason domain.CancelReason) (*domain.AgendaEvent, error) {
	var event *domain.AgendaEvent
	err := s.repo.Tx(ctx, func(ctx context.Context) error {
		e, err := s.repo.FindAgendaEvent(ctx, id)
		if err != nil || e == nil {
			event = e
			return err
		}
		e.Cancel(reason)
		event = e
		return s.repo.SaveAgendaEvent(ctx, e)
	})
	return event, err
}

func (s *AppointmentService) ProcessReminderTimesUp(ctx context.Context, eventID string) (*domain.AgendaEvent, error) {
	var event *domain.AgendaEvent
	err := s.repo.Tx(ctx, func(ctx context.Context) error {
		e, err := s.repo.FindAgendaEvent(ctx, eventID)
		if err != nil {
			return err
		}
		if e == nil {
			return fmt.Errorf("No event with id %s found", eventID)
		}
		e.MarkReminderSentRequested(s.clock.Now())
		event = e
		return s.repo.SaveAgendaEvent(ctx, e)
	})
	return event, err
}

func (s *AppointmentService) TrackPendingNotification(ctx context.Context, correlationKey, agendaEventID, notificationType string) error {
	return s.repo.SavePendingNotification(ctx, PendingNotification{
		CorrelationKey: correlationKey,
		AgendaEventID:  agendaEventID,
		Type:           notificationType,
	})
}

func (s *AppointmentService) ConfirmNotification(ctx context.Context, correlationKey string) (*domain.AgendaEvent, error) {
	var event *domain.AgendaEvent
	err := s.repo.Tx(ctx, func(ctx context.Context) error {
		pending, err := s.repo.FindPendingNotification(ctx, correlationKey)
		if err != nil || pending == nil {
			return err
		}
		if pending.Type != "reminder" && pending.Type != "Reminder" {
			return s.repo.RemovePendingNotification(ctx, correlationKey)
		}
		e, err := s.repo.FindAgendaEvent(ctx, pending.AgendaEventID)
		if err != nil {
			return err
		}
		if e == nil {
			return fmt.Errorf("No event with id %s found", pending.AgendaEventID)
		}
		e.MarkReminderSent(s.clock.Now())
		event = e
		if err := s.repo.SaveAgendaEvent(ctx, e); err != nil {
			return err
		}
		return s.repo.RemovePendingNotification(ctx, correlationKey)
	})
	return event, err
}

func (s *AppointmentService) saveAgenda(ctx context.Context, event *domain.AgendaEvent) error {
	return s.repo.Tx(ctx, func(ctx context.Context) error {
		return s.repo.SaveAgendaEvent(ctx, event)
	})
}

func (s *AppointmentService) resolveAttendee(ctx context.Context, typ domain.EventType, attendee domain.Attendee) (domain.Attendee, error) {
	if typ != domain.EventTypeAppointment {
		return domain.Attendee{ID: attendee.ID, DisplayName: "self"}, nil
	}
	customer, err := s.customerRegistry.FindByCustomerID(ctx, attendee.ID)
	if err != nil {
		return domain.Attendee{}, fmt.Errorf("error getting customer: %w", err)
	}
	if customer == nil {
		return domain.Attendee{}, fmt.Errorf("Unknown attendee %s", attendee.ID)
	}
	return domain.Attendee{ID: customer.ID, DisplayName: customer.DisplayName}, nil
}
