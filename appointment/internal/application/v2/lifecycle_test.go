package v2

import (
	"context"
	"testing"
)

type lifecycleEventHandlerStub struct {
	calls     int
	eventType string
	eventID   string
}

func (h *lifecycleEventHandlerStub) Handle(_ context.Context, eventType string, eventID string) error {
	h.calls++
	h.eventType = eventType
	h.eventID = eventID
	return nil
}

func TestCalendarLifecycleRoutesCalendarEventsToV2(t *testing.T) {
	calendar := &lifecycleEventHandlerStub{}
	legacy := &lifecycleEventHandlerStub{}
	handler := NewCalendarLifecycleHandler(calendar, legacy)

	if err := handler.Handle(context.Background(), "CalendarEventCreated", "event-1"); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if calendar.calls != 1 || legacy.calls != 0 || calendar.eventID != "event-1" {
		t.Fatalf("unexpected dispatch: calendar=%d legacy=%d", calendar.calls, legacy.calls)
	}
}

func TestCalendarLifecycleRoutesLegacyEventsToLegacyHandler(t *testing.T) {
	calendar := &lifecycleEventHandlerStub{}
	legacy := &lifecycleEventHandlerStub{}
	handler := NewCalendarLifecycleHandler(calendar, legacy)

	if err := handler.Handle(context.Background(), "AgendaEventScheduled", "event-1"); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if calendar.calls != 0 || legacy.calls != 1 {
		t.Fatalf("unexpected dispatch: calendar=%d legacy=%d", calendar.calls, legacy.calls)
	}
}

func TestCalendarLifecycleIgnoresUnknownEvents(t *testing.T) {
	calendar := &lifecycleEventHandlerStub{}
	legacy := &lifecycleEventHandlerStub{}
	handler := NewCalendarLifecycleHandler(calendar, legacy)

	if err := handler.Handle(context.Background(), "UnknownEvent", "event-1"); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if calendar.calls != 0 || legacy.calls != 0 {
		t.Fatalf("unexpected dispatch: calendar=%d legacy=%d", calendar.calls, legacy.calls)
	}
}
