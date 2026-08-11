package v2

type LifecycleEvent struct {
	Type            string `json:"type"`
	CalendarEventID string `json:"calendarEventId"`
}

func CalendarEventCreated(calendarEventID string) LifecycleEvent {
	return LifecycleEvent{Type: "CalendarEventCreated", CalendarEventID: calendarEventID}
}

func CalendarEventRescheduled(calendarEventID string) LifecycleEvent {
	return LifecycleEvent{Type: "CalendarEventRescheduled", CalendarEventID: calendarEventID}
}

func CalendarEventCanceled(calendarEventID string) LifecycleEvent {
	return LifecycleEvent{Type: "CalendarEventCanceled", CalendarEventID: calendarEventID}
}
