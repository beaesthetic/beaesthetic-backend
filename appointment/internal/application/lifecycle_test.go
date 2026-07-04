package application

import (
	"testing"
	"time"
)

func TestComputeReminderSendAtReturnsPotentialDateWhenFarEnough(t *testing.T) {
	now := time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)
	eventAt := now.Add(24 * time.Hour)
	sendAt, ok := computeReminderSendAt(now, eventAt, 2*time.Hour, 30*time.Minute, 2*time.Minute)
	if !ok {
		t.Fatal("expected reminder to be schedulable")
	}
	if sendAt == nil {
		t.Fatal("expected sendAt to be set")
	}
	want := eventAt.Add(-2 * time.Hour)
	if !sendAt.Equal(want) {
		t.Fatalf("sendAt=%s want=%s", sendAt, want)
	}
}

func TestComputeReminderSendAtReturnsImmediateDateWhenLateButStillSendable(t *testing.T) {
	now := time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)
	eventAt := now.Add(90 * time.Minute)
	sendAt, ok := computeReminderSendAt(now, eventAt, 2*time.Hour, 30*time.Minute, 2*time.Minute)
	if !ok {
		t.Fatal("expected reminder to be schedulable")
	}
	if sendAt == nil {
		t.Fatal("expected sendAt to be set")
	}
	want := now.Add(2 * time.Minute)
	if !sendAt.Equal(want) {
		t.Fatalf("sendAt=%s want=%s", sendAt, want)
	}
}

func TestComputeReminderSendAtReturnsUnprocessableWhenTooLate(t *testing.T) {
	now := time.Date(2026, time.July, 4, 12, 0, 0, 0, time.UTC)
	eventAt := now.Add(20 * time.Minute)
	sendAt, ok := computeReminderSendAt(now, eventAt, 2*time.Hour, 30*time.Minute, 2*time.Minute)
	if ok {
		t.Fatal("expected reminder to be unprocessable")
	}
	if sendAt != nil {
		t.Fatalf("expected nil sendAt, got %s", sendAt)
	}
}
