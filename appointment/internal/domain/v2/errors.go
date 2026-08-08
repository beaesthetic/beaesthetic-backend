package v2

import "errors"

var (
	ErrMissingRequiredData = errors.New("missing required data")
	ErrInvalidTimeRange    = errors.New("end must be after start")
	ErrInvalidCalendarID   = errors.New("invalid calendar id")
	ErrInvalidEventType    = errors.New("invalid agenda event type")
	ErrInvalidEventDetail  = errors.New("invalid agenda event detail")
	ErrInvalidVisibility   = errors.New("invalid agenda event visibility")
	ErrInvalidReminder     = errors.New("invalid appointment reminder")
	ErrInvalidNotification = errors.New("invalid appointment notification")
)
