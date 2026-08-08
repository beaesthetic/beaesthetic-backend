package v2

import domain "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"

type CalendarEventView struct {
	Event    domain.CalendarEvent
	Reminder *domain.AppointmentReminder
}
