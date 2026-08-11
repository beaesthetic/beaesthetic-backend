package v2

import (
	"time"

	domain "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
)

type CreateEventCommand interface {
	createEventCommand()
}

type CreateAppointmentCommand struct {
	CalendarID   string
	Start        time.Time
	End          time.Time
	Timezone     string
	AllDay       bool
	Title        string
	Description  string
	Visibility   domain.Visibility
	CustomerID   string
	Services     []domain.ServiceItem
	RemindBefore time.Duration
}

func (CreateAppointmentCommand) createEventCommand() {}

type CreateManualEventCommand struct {
	CalendarID    string
	Start         time.Time
	End           time.Time
	Timezone      string
	AllDay        bool
	Title         string
	Description   string
	Visibility    domain.Visibility
	ManualTitle   string
	ManualDetails string
	Location      *string
}

func (CreateManualEventCommand) createEventCommand() {}

type CreateTimeBlockCommand struct {
	CalendarID  string
	Start       time.Time
	End         time.Time
	Timezone    string
	AllDay      bool
	Title       string
	Description string
	Visibility  domain.Visibility
	Reason      string
}

func (CreateTimeBlockCommand) createEventCommand() {}

type CancelEventCommand struct {
	CalendarEventID string
	Reason          domain.CancelReason
}

type ListCalendarEventsQuery struct {
	CalendarID string
	Start      *time.Time
	End        *time.Time
	CustomerID string
	EventTypes []domain.CalendarEventType
}

type UpdateEventCommand interface {
	updateEventCommand()
}

type CalendarEventChanges struct {
	TimeRange   *TimeRangeUpdate
	Title       *string
	Description *string
	Visibility  *domain.Visibility
}

type UpdateCalendarFieldsCommand struct {
	CalendarEventID string
	Changes         CalendarEventChanges
}

func (UpdateCalendarFieldsCommand) updateEventCommand() {}

type UpdateAppointmentCommand struct {
	CalendarEventID string
	Changes         CalendarEventChanges
	Services        []domain.ServiceItem
}

func (UpdateAppointmentCommand) updateEventCommand() {}

type UpdateManualEventCommand struct {
	CalendarEventID string
	Changes         CalendarEventChanges
	Title           *string
	Description     *string
	Location        **string
}

func (UpdateManualEventCommand) updateEventCommand() {}

type UpdateTimeBlockCommand struct {
	CalendarEventID string
	Changes         CalendarEventChanges
	Reason          string
}

func (UpdateTimeBlockCommand) updateEventCommand() {}

type TimeRangeUpdate struct {
	Start    time.Time
	End      time.Time
	Timezone string
	AllDay   bool
}
