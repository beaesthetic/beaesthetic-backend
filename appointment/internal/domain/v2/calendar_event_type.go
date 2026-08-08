package v2

import "time"

type CalendarEventType string

const (
	CalendarEventTypeAppointment CalendarEventType = "appointment"
	CalendarEventTypeManual      CalendarEventType = "manual"
	CalendarEventTypeTimeBlock   CalendarEventType = "time_block"
)

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

type CancelReason string

const (
	CancelReasonCustomer CancelReason = "customer_cancel"
	CancelReasonDeleted  CancelReason = "deleted"
)

type CalendarEventCancellation struct {
	Reason     CancelReason
	CanceledAt time.Time
}

type TimeRange struct {
	Start    time.Time
	End      time.Time
	Timezone string
	AllDay   bool
}

func NewTimeRange(start time.Time, end time.Time, timezone string, allDay bool) (TimeRange, error) {
	if timezone == "" {
		timezone = "UTC"
	}
	if !end.After(start) {
		return TimeRange{}, ErrInvalidTimeRange
	}
	return TimeRange{Start: start.UTC(), End: end.UTC(), Timezone: timezone, AllDay: allDay}, nil
}

func (eventRange TimeRange) Equals(other TimeRange) bool {
	return eventRange.Start.Equal(other.Start) &&
		eventRange.End.Equal(other.End) &&
		eventRange.Timezone == other.Timezone &&
		eventRange.AllDay == other.AllDay
}

func (eventType CalendarEventType) Valid() bool {
	switch eventType {
	case CalendarEventTypeAppointment, CalendarEventTypeManual, CalendarEventTypeTimeBlock:
		return true
	default:
		return false
	}
}

func (visibility Visibility) Valid() bool {
	switch visibility {
	case VisibilityPublic, VisibilityPrivate:
		return true
	default:
		return false
	}
}
