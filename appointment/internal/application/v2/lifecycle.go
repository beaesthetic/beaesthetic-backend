package v2

import "context"

type LifecycleEventHandler interface {
	Handle(ctx context.Context, eventType string, calendarEventID string) error
}

type CalendarLifecycleHandler struct {
	calendar LifecycleEventHandler
	legacy   LifecycleEventHandler
}

func NewCalendarLifecycleHandler(calendar LifecycleEventHandler, legacy LifecycleEventHandler) *CalendarLifecycleHandler {
	return &CalendarLifecycleHandler{calendar: calendar, legacy: legacy}
}

func (h *CalendarLifecycleHandler) Handle(ctx context.Context, eventType string, calendarEventID string) error {
	if calendarEventID == "" {
		return nil
	}
	if isLegacyAppointmentLifecycleEvent(eventType) {
		if h.legacy == nil {
			return nil
		}
		return h.legacy.Handle(ctx, eventType, calendarEventID)
	}
	if !isCalendarLifecycleEvent(eventType) || h.calendar == nil {
		return nil
	}
	return h.calendar.Handle(ctx, eventType, calendarEventID)
}

func isLegacyAppointmentLifecycleEvent(eventType string) bool {
	switch eventType {
	case "AgendaEventScheduled", "AgendaEventCreated", "AgendaEventRescheduled", "AgendaEventDeleted", "AgendaEventCanceled":
		return true
	default:
		return false
	}
}

func isCalendarLifecycleEvent(eventType string) bool {
	switch eventType {
	case "CalendarEventCreated", "CalendarEventRescheduled", "CalendarEventCanceled":
		return true
	default:
		return false
	}
}
