package domain

type EventType string

const (
	EventTypeAppointment EventType = "appointment"
	EventTypeGeneric     EventType = "event"
)

type CancelReason string

const (
	CancelReasonCustomer CancelReason = "customer_cancel"
	CancelReasonDeleted  CancelReason = "deleted"
)
