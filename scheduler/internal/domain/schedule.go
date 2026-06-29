package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ScheduleJob struct {
	ID         uuid.UUID
	Route      string
	Payload    map[string]any
	ScheduleAt time.Time
	Attempts   int
	LastError  *string
}

type JobRepository interface {
	Save(ctx context.Context, job ScheduleJob) error
	Delete(ctx context.Context, id uuid.UUID) error
	PollDue(ctx context.Context, now time.Time, batchSize int, leaseTTL time.Duration, leaseOwner string) ([]ScheduleJob, error)
	Ack(ctx context.Context, id uuid.UUID) error
	Nack(ctx context.Context, id uuid.UUID, reason string) error
}

type Publisher interface {
	Publish(ctx context.Context, job ScheduleJob) error
}
