package v2

import "time"

type NotificationKind string

const (
	NotificationKindConfirmation NotificationKind = "confirmation"
	NotificationKindRescheduled  NotificationKind = "rescheduled"
	NotificationKindReminder     NotificationKind = "reminder"
)

type NotificationType string

const (
	NotificationTypeAppointmentConfirmation NotificationType = "appointment_confirmation"
	NotificationTypeAppointmentRescheduled  NotificationType = "appointment_rescheduled"
	NotificationTypeAppointmentReminder     NotificationType = "appointment_reminder"
)

type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
	NotificationStatusExpired NotificationStatus = "expired"
)

const customerNotificationRecipientKind = "customer"

type NotificationRecipient struct {
	kind string
	id   string
}

func NewCustomerNotificationRecipient(customerID string) (NotificationRecipient, error) {
	return newNotificationRecipient(customerNotificationRecipientKind, customerID)
}

func ReconstituteNotificationRecipient(kind string, id string) (NotificationRecipient, error) {
	return newNotificationRecipient(kind, id)
}

func newNotificationRecipient(kind string, id string) (NotificationRecipient, error) {
	if kind == "" || id == "" {
		return NotificationRecipient{}, ErrInvalidNotification
	}
	return NotificationRecipient{kind: kind, id: id}, nil
}

func (recipient NotificationRecipient) Kind() string {
	return recipient.kind
}

func (recipient NotificationRecipient) ID() string {
	return recipient.id
}

type AppointmentNotification struct {
	CorrelationKey  string
	CalendarEventID string
	Kind            NotificationKind
	Type            NotificationType
	Status          NotificationStatus
	Recipient       NotificationRecipient
	IdempotencyKey  *string
	FailureReason   *string
	FailureMessage  *string
	CreatedAt       time.Time
	CompletedAt     *time.Time
	ExpiresAt       time.Time
}

func NewAppointmentNotification(correlationKey string, calendarEventID string, kind NotificationKind, recipient NotificationRecipient, idempotencyKey *string, createdAt time.Time, expiresAt time.Time) (AppointmentNotification, error) {
	if correlationKey == "" || calendarEventID == "" || recipient.id == "" || recipient.kind == "" || !expiresAt.After(createdAt) {
		return AppointmentNotification{}, ErrInvalidNotification
	}
	notificationType, ok := notificationTypeForKind(kind)
	if !ok {
		return AppointmentNotification{}, ErrInvalidNotification
	}
	return AppointmentNotification{
		CorrelationKey:  correlationKey,
		CalendarEventID: calendarEventID,
		Kind:            kind,
		Type:            notificationType,
		Status:          NotificationStatusPending,
		Recipient:       recipient,
		IdempotencyKey:  idempotencyKey,
		CreatedAt:       createdAt.UTC(),
		ExpiresAt:       expiresAt.UTC(),
	}, nil
}

func ReconstituteAppointmentNotification(notification AppointmentNotification) (AppointmentNotification, error) {
	if notification.CorrelationKey == "" || notification.CalendarEventID == "" ||
		notification.Recipient.id == "" || notification.Recipient.kind == "" ||
		!notification.Status.Valid() || !notification.ExpiresAt.After(notification.CreatedAt) {
		return AppointmentNotification{}, ErrInvalidNotification
	}
	if notificationType, ok := notificationTypeForKind(notification.Kind); !ok || notification.Type != notificationType {
		return AppointmentNotification{}, ErrInvalidNotification
	}
	notification.CreatedAt = notification.CreatedAt.UTC()
	notification.CompletedAt = utcTimePointer(notification.CompletedAt)
	notification.ExpiresAt = notification.ExpiresAt.UTC()
	return notification, nil
}

func (status NotificationStatus) Valid() bool {
	switch status {
	case NotificationStatusPending, NotificationStatusSent, NotificationStatusFailed, NotificationStatusExpired:
		return true
	default:
		return false
	}
}

func (notification *AppointmentNotification) MarkSent(now time.Time) {
	completedAt := now.UTC()
	notification.Status = NotificationStatusSent
	notification.FailureReason = nil
	notification.FailureMessage = nil
	notification.CompletedAt = &completedAt
}

func (notification *AppointmentNotification) MarkFailed(reason string, message string, now time.Time) {
	completedAt := now.UTC()
	notification.Status = NotificationStatusFailed
	notification.FailureReason = optionalString(reason)
	notification.FailureMessage = optionalString(message)
	notification.CompletedAt = &completedAt
}

func (notification *AppointmentNotification) MarkExpired(now time.Time) {
	completedAt := now.UTC()
	notification.Status = NotificationStatusExpired
	notification.CompletedAt = &completedAt
}

func notificationTypeForKind(kind NotificationKind) (NotificationType, bool) {
	switch kind {
	case NotificationKindConfirmation:
		return NotificationTypeAppointmentConfirmation, true
	case NotificationKindRescheduled:
		return NotificationTypeAppointmentRescheduled, true
	case NotificationKindReminder:
		return NotificationTypeAppointmentReminder, true
	default:
		return "", false
	}
}
