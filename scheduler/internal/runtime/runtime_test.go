package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/beaesthetic/scheduler/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestTickAcksAfterSuccessfulPublish(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepository{jobs: []domain.ScheduleJob{{ID: id, Route: "route", Payload: map[string]any{"ok": true}}}}
	publisher := &fakePublisher{}
	r := New(repo, publisher, zap.NewNop(), time.Second, time.Second, 10)

	r.tick(context.Background())

	if publisher.published != 1 {
		t.Fatalf("published = %d", publisher.published)
	}
	if repo.acked != id {
		t.Fatalf("acked = %s", repo.acked)
	}
	if repo.nacked != uuid.Nil {
		t.Fatalf("nacked = %s", repo.nacked)
	}
}

func TestTickNacksAfterPublishFailure(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepository{jobs: []domain.ScheduleJob{{ID: id, Route: "route", Payload: map[string]any{"ok": true}}}}
	publisher := &fakePublisher{err: errors.New("publish failed")}
	r := New(repo, publisher, zap.NewNop(), time.Second, time.Second, 10)

	r.tick(context.Background())

	if repo.acked != uuid.Nil {
		t.Fatalf("acked = %s", repo.acked)
	}
	if repo.nacked != id {
		t.Fatalf("nacked = %s", repo.nacked)
	}
}

type fakeRepository struct {
	jobs   []domain.ScheduleJob
	acked  uuid.UUID
	nacked uuid.UUID
}

func (f *fakeRepository) Save(context.Context, domain.ScheduleJob) error { return nil }
func (f *fakeRepository) Delete(context.Context, uuid.UUID) error        { return nil }
func (f *fakeRepository) PollDue(context.Context, time.Time, int, time.Duration, string) ([]domain.ScheduleJob, error) {
	return f.jobs, nil
}
func (f *fakeRepository) Ack(_ context.Context, id uuid.UUID) error {
	f.acked = id
	return nil
}
func (f *fakeRepository) Nack(_ context.Context, id uuid.UUID, _ string) error {
	f.nacked = id
	return nil
}

type fakePublisher struct {
	published int
	err       error
}

func (f *fakePublisher) Publish(context.Context, domain.ScheduleJob) error {
	f.published++
	return f.err
}
