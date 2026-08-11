package v2

import (
	"strings"
	"time"
)

const DefaultCalendarID = "d2a36e25-4824-4167-a062-a5af96f97703"

type EventDetail interface {
	EventType() CalendarEventType
}

type CalendarEvent struct {
	ID           string
	CalendarID   string
	Type         CalendarEventType
	Range        TimeRange
	Title        string
	Description  string
	Visibility   Visibility
	Detail       EventDetail
	Cancellation *CalendarEventCancellation
	Version      int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
	events       []LifecycleEvent
}

type AppointmentEventParams struct {
	EventID     string
	CalendarID  string
	Range       TimeRange
	Title       string
	Description string
	Visibility  Visibility
	Customer    CustomerRef
	Services    []ServiceItem
	Now         time.Time
}

func NewAppointmentEvent(params AppointmentEventParams) (CalendarEvent, error) {
	detail, err := NewAppointment(params.Customer, params.Services, params.Now)
	if err != nil {
		return CalendarEvent{}, err
	}
	return newCalendarEvent(params.EventID, params.CalendarID, params.Range, params.Title, params.Description, params.Visibility, detail, params.Now)
}

type ManualEventParams struct {
	EventID          string
	CalendarID       string
	Range            TimeRange
	EventTitle       string
	EventDescription string
	Visibility       Visibility
	Title            string
	Description      string
	Location         *string
	Now              time.Time
}

func NewManualCalendarEvent(params ManualEventParams) (CalendarEvent, error) {
	detail, err := NewManualEvent(params.Title, params.Description, params.Location)
	if err != nil {
		return CalendarEvent{}, err
	}
	return newCalendarEvent(params.EventID, params.CalendarID, params.Range, params.EventTitle, params.EventDescription, params.Visibility, detail, params.Now)
}

type TimeBlockEventParams struct {
	EventID     string
	CalendarID  string
	Range       TimeRange
	Title       string
	Description string
	Visibility  Visibility
	Reason      string
	Now         time.Time
}

func NewTimeBlockCalendarEvent(params TimeBlockEventParams) (CalendarEvent, error) {
	detail, err := NewTimeBlock(params.Reason)
	if err != nil {
		return CalendarEvent{}, err
	}
	return newCalendarEvent(params.EventID, params.CalendarID, params.Range, params.Title, params.Description, params.Visibility, detail, params.Now)
}

func ReconstituteCalendarEvent(event CalendarEvent) (CalendarEvent, error) {
	if event.ID == "" || event.Detail == nil {
		return CalendarEvent{}, ErrMissingRequiredData
	}
	calendarID, err := NormalizeCalendarID(event.CalendarID)
	if err != nil {
		return CalendarEvent{}, err
	}
	event.CalendarID = calendarID
	if !event.Type.Valid() || event.Type != event.Detail.EventType() {
		return CalendarEvent{}, ErrInvalidEventType
	}
	if event.Visibility == "" {
		event.Visibility = VisibilityPrivate
	}
	if !event.Visibility.Valid() {
		return CalendarEvent{}, ErrInvalidVisibility
	}
	return event, nil
}

func newCalendarEvent(id string, calendarID string, eventRange TimeRange, title string, description string, visibility Visibility, detail EventDetail, now time.Time) (CalendarEvent, error) {
	if id == "" || detail == nil {
		return CalendarEvent{}, ErrMissingRequiredData
	}
	normalizedCalendarID, err := NormalizeCalendarID(calendarID)
	if err != nil {
		return CalendarEvent{}, err
	}
	eventType := detail.EventType()
	if !eventType.Valid() {
		return CalendarEvent{}, ErrInvalidEventType
	}
	if visibility == "" {
		visibility = VisibilityPrivate
	}
	if !visibility.Valid() {
		return CalendarEvent{}, ErrInvalidVisibility
	}
	event := CalendarEvent{
		ID:          id,
		CalendarID:  normalizedCalendarID,
		Type:        eventType,
		Range:       eventRange,
		Title:       title,
		Description: description,
		Visibility:  visibility,
		Detail:      detail,
		Version:     1,
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC(),
	}
	event.record(CalendarEventCreated(event.ID))
	return event, nil
}

func NormalizeCalendarID(calendarID string) (string, error) {
	calendarID = strings.TrimSpace(calendarID)
	if calendarID == "" || strings.EqualFold(calendarID, DefaultCalendarID) {
		return DefaultCalendarID, nil
	}
	return "", ErrInvalidCalendarID
}

func (event *CalendarEvent) Reschedule(eventRange TimeRange, now time.Time) {
	if event.Range.Equals(eventRange) {
		return
	}
	event.Range = eventRange
	event.UpdatedAt = now.UTC()
	event.record(CalendarEventRescheduled(event.ID))
}

func (event *CalendarEvent) ChangeTitle(title string, now time.Time) {
	event.Title = title
	event.UpdatedAt = now.UTC()
}

func (event *CalendarEvent) ChangeDescription(description string, now time.Time) {
	event.Description = description
	event.UpdatedAt = now.UTC()
}

func (event *CalendarEvent) ChangeVisibility(visibility Visibility, now time.Time) error {
	if visibility == "" {
		visibility = VisibilityPrivate
	}
	if !visibility.Valid() {
		return ErrInvalidVisibility
	}
	event.Visibility = visibility
	event.UpdatedAt = now.UTC()
	return nil
}

func (event *CalendarEvent) ReplaceAppointmentServices(services []ServiceItem, now time.Time) error {
	appointment, ok := event.Detail.(Appointment)
	if !ok {
		return ErrInvalidEventDetail
	}
	appointment.ReplaceServices(services, now)
	event.Detail = appointment
	event.UpdatedAt = now.UTC()
	return nil
}

func (event *CalendarEvent) ChangeManualDetails(title string, description string, location *string, now time.Time) error {
	manualEvent, ok := event.Detail.(ManualEvent)
	if !ok {
		return ErrInvalidEventDetail
	}
	if err := manualEvent.Rename(title); err != nil {
		return err
	}
	manualEvent.ChangeDescription(description)
	manualEvent.ChangeLocation(location)
	event.Detail = manualEvent
	event.UpdatedAt = now.UTC()
	return nil
}

func (event *CalendarEvent) ChangeTimeBlockReason(reason string, now time.Time) error {
	timeBlock, ok := event.Detail.(TimeBlock)
	if !ok {
		return ErrInvalidEventDetail
	}
	if err := timeBlock.ChangeReason(reason); err != nil {
		return err
	}
	event.Detail = timeBlock
	event.UpdatedAt = now.UTC()
	return nil
}

func (event *CalendarEvent) Cancel(reason CancelReason, now time.Time) {
	if event.IsCanceled() {
		return
	}
	event.Cancellation = &CalendarEventCancellation{
		Reason:     reason,
		CanceledAt: now.UTC(),
	}
	event.UpdatedAt = now.UTC()
	event.record(CalendarEventCanceled(event.ID))
}

func (event CalendarEvent) IsCanceled() bool {
	return event.Cancellation != nil
}

func (event *CalendarEvent) PullEvents() []LifecycleEvent {
	pulled := event.events
	event.events = nil
	return pulled
}

func (event *CalendarEvent) record(lifecycleEvent LifecycleEvent) {
	event.events = append(event.events, lifecycleEvent)
}
