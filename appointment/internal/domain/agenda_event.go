package domain

import "time"

type AgendaEvent struct {
	ID             string
	Type           EventType
	Title          string
	Description    string
	Start          time.Time
	End            time.Time
	Attendee       Attendee
	Services       []AppointmentServiceRef
	CancelReason   *CancelReason
	ReminderStatus ReminderStatus
	ReminderSentAt *time.Time
	RemindBefore   time.Duration
	Version        int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	events         []LifecycleEvent
}

func NewAgendaEvent(
	id string,
	typ EventType,
	title string,
	description string,
	start time.Time,
	end time.Time,
	attendee Attendee,
	services []AppointmentServiceRef,
	remindBefore time.Duration,
	now time.Time,
) (AgendaEvent, error) {
	if id == "" || attendee.ID == "" {
		return AgendaEvent{}, ErrMissingRequiredAgendaData
	}
	if !end.After(start) {
		return AgendaEvent{}, ErrInvalidTimeSpan
	}

	event := AgendaEvent{
		ID:             id,
		Type:           typ,
		Title:          title,
		Description:    description,
		Start:          start.UTC(),
		End:            end.UTC(),
		Attendee:       attendee,
		Services:       services,
		ReminderStatus: ReminderPending,
		RemindBefore:   remindBefore,
		CreatedAt:      now.UTC(),
		UpdatedAt:      now.UTC(),
	}
	event.record(AgendaEventScheduled(event.ID))
	return event, nil
}

func (event *AgendaEvent) Reschedule(start time.Time, end time.Time) error {
	if !end.After(start) {
		return ErrInvalidTimeSpan
	}
	event.Start = start.UTC()
	event.End = end.UTC()
	event.UpdatedAt = time.Now().UTC()
	event.record(AgendaEventRescheduled(event.ID))
	return nil
}

func (event *AgendaEvent) Cancel(reason CancelReason) {
	event.CancelReason = &reason
	event.UpdatedAt = time.Now().UTC()
	event.record(AgendaEventDeleted(event.ID))
}

func (event *AgendaEvent) PullEvents() []LifecycleEvent {
	events := event.events
	event.events = nil
	return events
}

func (event *AgendaEvent) record(lifecycleEvent LifecycleEvent) {
	event.events = append(event.events, lifecycleEvent)
}

func (event *AgendaEvent) MarkReminderSentRequested(now time.Time) {
	event.ReminderStatus = ReminderSentRequested
	event.UpdatedAt = now.UTC()
}

func (event *AgendaEvent) MarkReminderSent(now time.Time) {
	event.ReminderStatus = ReminderSent
	sentAt := now.UTC()
	event.ReminderSentAt = &sentAt
	event.UpdatedAt = sentAt
}
