package runtime

import (
	"context"
	"time"

	"github.com/beaesthetic/scheduler/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Runtime struct {
	repository      domain.JobRepository
	publisher       domain.Publisher
	logger          *zap.Logger
	pollingInterval time.Duration
	leaseTTL        time.Duration
	batchSize       int
	leaseOwner      string
}

func New(repository domain.JobRepository, publisher domain.Publisher, logger *zap.Logger, pollingInterval time.Duration, leaseTTL time.Duration, batchSize int) *Runtime {
	return &Runtime{
		repository:      repository,
		publisher:       publisher,
		logger:          logger,
		pollingInterval: pollingInterval,
		leaseTTL:        leaseTTL,
		batchSize:       batchSize,
		leaseOwner:      uuid.NewString(),
	}
}

func (r *Runtime) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.pollingInterval)
	defer ticker.Stop()

	r.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}

func (r *Runtime) tick(ctx context.Context) {
	start := time.Now()
	jobs, err := r.repository.PollDue(ctx, start, r.batchSize, r.leaseTTL, r.leaseOwner)
	if err != nil {
		r.logger.Error("failed to poll due jobs", zap.Error(err))
		return
	}

	if len(jobs) > 0 {
		r.logger.Info("polled due jobs", zap.Int("count", len(jobs)), zap.Duration("duration", time.Since(start)))
	}

	for _, job := range jobs {
		fields := []zap.Field{
			zap.String("schedule_id", job.ID.String()),
			zap.String("route", job.Route),
			zap.Int("attempts", job.Attempts),
			zap.String("lease_owner", r.leaseOwner),
		}

		if err := r.publisher.Publish(ctx, job); err != nil {
			r.logger.Error("failed to publish schedule job", append(fields, zap.Error(err))...)
			_ = r.repository.Nack(ctx, job.ID, err.Error())
			continue
		}

		if err := r.repository.Ack(ctx, job.ID); err != nil {
			r.logger.Error("failed to ack schedule job", append(fields, zap.Error(err))...)
			continue
		}

		r.logger.Info("published schedule job", fields...)
	}
}
