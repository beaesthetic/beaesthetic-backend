package application

import (
	"context"
	"time"

	"github.com/beaesthetic/scheduler/internal/domain"
	"github.com/google/uuid"
)

type SchedulerService struct {
	repository domain.JobRepository
}

func NewSchedulerService(repository domain.JobRepository) *SchedulerService {
	return &SchedulerService{repository: repository}
}

func (s *SchedulerService) Schedule(ctx context.Context, id uuid.UUID, scheduleAt time.Time, route string, payload map[string]any) error {
	return s.repository.Save(ctx, domain.ScheduleJob{
		ID:         id,
		Route:      route,
		Payload:    payload,
		ScheduleAt: scheduleAt,
	})
}

func (s *SchedulerService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
