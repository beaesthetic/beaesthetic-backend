package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

type ServiceRepository interface {
	SaveService(context.Context, domain.AppointmentService) (domain.AppointmentService, error)
	FindServices(context.Context) ([]domain.AppointmentService, error)
	SearchServices(context.Context, string, int) ([]domain.AppointmentService, error)
	FindService(context.Context, string) (*domain.AppointmentService, error)
}

type ServiceService struct {
	repo ServiceRepository
}

func NewServiceService(repo ServiceRepository) *ServiceService {
	return &ServiceService{repo: repo}
}

func (s *ServiceService) CreateService(ctx context.Context, name string, price float64, tags []string, color *string) (domain.AppointmentService, error) {
	return s.repo.SaveService(ctx, domain.AppointmentService{ID: uuid.NewString(), Name: name, Price: price, Tags: tags, Color: color})
}

func (s *ServiceService) UpdateService(ctx context.Context, id string, price *float64, tags []string, color *string) (domain.AppointmentService, error) {
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

func (s *ServiceService) AllServices(ctx context.Context) ([]domain.AppointmentService, error) {
	return s.repo.FindServices(ctx)
}

func (s *ServiceService) FindService(ctx context.Context, id string) (*domain.AppointmentService, error) {
	return s.repo.FindService(ctx, id)
}

func (s *ServiceService) SearchServices(ctx context.Context, text string, limit int) ([]domain.AppointmentService, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.SearchServices(ctx, text, limit)
}
