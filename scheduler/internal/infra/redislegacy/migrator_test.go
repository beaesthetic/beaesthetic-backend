package redislegacy

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestParseLegacyJob(t *testing.T) {
	id := uuid.New()
	raw := `{"id":{"id":"` + id.String() + `"},"meta":{"route":"reminders","data":{"kind":"reminder","count":2}},"scheduleAt":"2026-06-27T10:00:00Z"}`

	job, err := parseLegacyJob(raw)
	if err != nil {
		t.Fatalf("parseLegacyJob() error = %v", err)
	}

	if job.ID != id {
		t.Fatalf("ID = %s", job.ID)
	}
	if job.Route != "reminders" {
		t.Fatalf("Route = %q", job.Route)
	}
	if job.Payload["kind"] != "reminder" {
		t.Fatalf("Payload[kind] = %v", job.Payload["kind"])
	}
	if !job.ScheduleAt.Equal(time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("ScheduleAt = %s", job.ScheduleAt)
	}
}

func TestParseLegacyJobWithStringID(t *testing.T) {
	id := uuid.New()
	raw := `{"id":"` + id.String() + `","meta":{"route":"reminders","data":{}},"scheduleAt":"2026-06-27T10:00:00Z"}`

	job, err := parseLegacyJob(raw)
	if err != nil {
		t.Fatalf("parseLegacyJob() error = %v", err)
	}

	if job.ID != id {
		t.Fatalf("ID = %s", job.ID)
	}
}
