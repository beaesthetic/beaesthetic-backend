package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

type Repository interface {
	SaveAgendaEvent(context.Context, *domain.AgendaEvent) error
	FindAgendaEvent(context.Context, string) (*domain.AgendaEvent, error)
	SearchAgendaEvents(context.Context, string, *time.Time, *time.Time) ([]domain.AgendaEvent, error)
	SaveService(context.Context, domain.AppointmentService) (domain.AppointmentService, error)
	FindServices(context.Context) ([]domain.AppointmentService, error)
	SearchServices(context.Context, string, int) ([]domain.AppointmentService, error)
	FindService(context.Context, string) (*domain.AppointmentService, error)
}

type Service struct {
	repo           Repository
	now            func() time.Time
	reminderBefore time.Duration
}

func NewService(repo Repository, reminderBefore time.Duration) *Service {
	return &Service{repo: repo, now: time.Now, reminderBefore: reminderBefore}
}

func (s *Service) CreateAgenda(ctx context.Context, typ domain.EventType, title, description string, start, end time.Time, attendee domain.Attendee, services []domain.AppointmentServiceRef) (*domain.AgendaEvent, error) {
	e, err := domain.NewAgendaEvent(uuid.NewString(), typ, title, description, start, end, attendee, services, s.reminderBefore, s.now())
	if err != nil {
		return nil, err
	}
	return &e, s.repo.SaveAgendaEvent(ctx, &e)
}
func (s *Service) GetAgenda(ctx context.Context, id string) (*domain.AgendaEvent, error) {
	return s.repo.FindAgendaEvent(ctx, id)
}
func (s *Service) SearchAgenda(ctx context.Context, attendee string, start, end *time.Time) ([]domain.AgendaEvent, error) {
	return s.repo.SearchAgendaEvents(ctx, attendee, start, end)
}
func (s *Service) UpdateAgenda(ctx context.Context, id string, title, description *string, start, end *time.Time, services []domain.AppointmentServiceRef) (*domain.AgendaEvent, error) {
	e, err := s.repo.FindAgendaEvent(ctx, id)
	if err != nil || e == nil {
		return e, err
	}
	if title != nil {
		e.Title = *title
	}
	if description != nil {
		e.Description = *description
	}
	if start != nil && end != nil {
		if err := e.Reschedule(*start, *end); err != nil {
			return nil, err
		}
	}
	if services != nil {
		e.Services = services
	}
	return e, s.repo.SaveAgendaEvent(ctx, e)
}
func (s *Service) DeleteAgenda(ctx context.Context, id string, reason domain.CancelReason) (*domain.AgendaEvent, error) {
	e, err := s.repo.FindAgendaEvent(ctx, id)
	if err != nil || e == nil {
		return e, err
	}
	e.Cancel(reason)
	return e, s.repo.SaveAgendaEvent(ctx, e)
}
func (s *Service) CreateService(ctx context.Context, name string, price float64, tags []string, color *string) (domain.AppointmentService, error) {
	return s.repo.SaveService(ctx, domain.AppointmentService{ID: uuid.NewString(), Name: name, Price: price, Tags: tags, Color: color})
}
func (s *Service) UpdateService(ctx context.Context, id string, price *float64, tags []string, color *string) (domain.AppointmentService, error) {
	svc, err := s.repo.FindService(ctx, id)
	if err != nil {
		return domain.AppointmentService{}, err
	}
	if svc == nil {
		return domain.AppointmentService{}, nil
	}
	if price != nil {
		svc.Price = *price
	}
	if tags != nil {
		svc.Tags = tags
	}
	if color != nil {
		svc.Color = color
	}
	return s.repo.SaveService(ctx, *svc)
}
func (s *Service) AllServices(ctx context.Context) ([]domain.AppointmentService, error) {
	return s.repo.FindServices(ctx)
}
func (s *Service) SearchServices(ctx context.Context, text string, limit int) ([]domain.AppointmentService, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.SearchServices(ctx, text, limit)
}
