package application

import (
	"context"
	"fmt"
	"strings"

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

func (s *ServiceService) CreateService(ctx context.Context, name string, tags []string, color *string) (domain.AppointmentService, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.AppointmentService{}, fmt.Errorf("service name is required")
	}
	return s.repo.SaveService(ctx, domain.AppointmentService{ID: uuid.NewString(), Name: name, Tags: tags, Color: color})
}

func (s *ServiceService) UpdateService(ctx context.Context, id string, tags []string, color *string) (*domain.AppointmentService, error) {
	service, err := s.repo.FindService(ctx, id)
	if err != nil || service == nil {
		return service, err
	}
	if tags != nil {
		service.Tags = tags
	}
	if color != nil {
		service.Color = color
	}
	updated, err := s.repo.SaveService(ctx, *service)
	if err != nil {
		return nil, err
	}
	return &updated, nil
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
	// The repository search also supports an empty query, preserving the limit for
	// an unfiltered catalog listing.
	return s.repo.SearchServices(ctx, text, limit)
}
