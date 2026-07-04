package domain

import "errors"

var (
	ErrMissingRequiredAgendaData = errors.New("id and attendee are required")
	ErrInvalidTimeSpan           = errors.New("end must be after start")
)
