package domain

import (
	"testing"
	"time"
)

func TestAgendaEventLifecycle(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour)
	e, err := NewAgendaEvent("id", EventTypeGeneric, "title", "desc", start, end, Attendee{ID: "attendee", DisplayName: "Self"}, nil, time.Hour, start)
	if err != nil {
		t.Fatal(err)
	}
	if len(e.PullEvents()) != 1 {
		t.Fatal("expected scheduled event")
	}
	if err := e.Reschedule(start.Add(time.Hour), end.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if len(e.PullEvents()) != 1 {
		t.Fatal("expected rescheduled event")
	}
	e.Cancel(CancelReasonDeleted)
	if e.CancelReason == nil {
		t.Fatal("expected cancel reason")
	}
}
