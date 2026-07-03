package domain

type LifecycleEvent struct {
	Type          string
	AgendaEventID string
}

func AgendaEventScheduled(agendaEventID string) LifecycleEvent {
	return LifecycleEvent{Type: "AgendaEventScheduled", AgendaEventID: agendaEventID}
}

func AgendaEventRescheduled(agendaEventID string) LifecycleEvent {
	return LifecycleEvent{Type: "AgendaEventRescheduled", AgendaEventID: agendaEventID}
}

func AgendaEventDeleted(agendaEventID string) LifecycleEvent {
	return LifecycleEvent{Type: "AgendaEventDeleted", AgendaEventID: agendaEventID}
}
