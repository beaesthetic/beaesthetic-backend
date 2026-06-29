package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/beaesthetic/scheduler/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepository struct {
	pool          *pgxpool.Pool
	schedulerName string
}

func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func NewJobRepository(pool *pgxpool.Pool, schedulerName string) *JobRepository {
	return &JobRepository{pool: pool, schedulerName: schedulerName}
}

func (r *JobRepository) Save(ctx context.Context, job domain.ScheduleJob) error {
	payload, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO schedule_jobs (id, scheduler_name, route, payload, schedule_at, leased_until, lease_owner, attempts, last_error, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NULL, NULL, 0, NULL, now())
		 ON CONFLICT (id) DO UPDATE
		 SET scheduler_name = EXCLUDED.scheduler_name,
		     route = EXCLUDED.route,
		     payload = EXCLUDED.payload,
		     schedule_at = EXCLUDED.schedule_at,
		     leased_until = NULL,
		     lease_owner = NULL,
		     last_error = NULL,
		     updated_at = now()`,
		job.ID, r.schedulerName, job.Route, payload, job.ScheduleAt.UTC(),
	)
	return err
}

func (r *JobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM schedule_jobs WHERE id = $1 AND scheduler_name = $2`, id, r.schedulerName)
	return err
}

func (r *JobRepository) PollDue(ctx context.Context, now time.Time, batchSize int, leaseTTL time.Duration, leaseOwner string) ([]domain.ScheduleJob, error) {
	leasedUntil := now.Add(leaseTTL)
	rows, err := r.pool.Query(ctx,
		`WITH due AS (
		   SELECT id
		   FROM schedule_jobs
		   WHERE scheduler_name = $1
		     AND schedule_at <= $2
		     AND (leased_until IS NULL OR leased_until <= $2)
		   ORDER BY schedule_at ASC
		   LIMIT $3
		   FOR UPDATE SKIP LOCKED
		 )
		 UPDATE schedule_jobs j
		 SET leased_until = $4,
		     lease_owner = $5,
		     attempts = attempts + 1,
		     updated_at = now()
		 FROM due
		 WHERE j.id = due.id
		 RETURNING j.id, j.route, j.payload, j.schedule_at, j.attempts, j.last_error`,
		r.schedulerName, now.UTC(), batchSize, leasedUntil.UTC(), leaseOwner,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]domain.ScheduleJob, 0)
	for rows.Next() {
		var job domain.ScheduleJob
		var payload []byte
		if err := rows.Scan(&job.ID, &job.Route, &payload, &job.ScheduleAt, &job.Attempts, &job.LastError); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(payload, &job.Payload); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (r *JobRepository) Ack(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM schedule_jobs WHERE id = $1 AND scheduler_name = $2`, id, r.schedulerName)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *JobRepository) Nack(ctx context.Context, id uuid.UUID, reason string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE schedule_jobs
		 SET leased_until = NULL,
		     lease_owner = NULL,
		     last_error = $3,
		     updated_at = now()
		 WHERE id = $1 AND scheduler_name = $2`,
		id, r.schedulerName, reason,
	)
	return err
}

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
