package v2

import "time"

type ReminderStatus string

const (
	ReminderStatusPending       ReminderStatus = "pending"
	ReminderStatusScheduled     ReminderStatus = "scheduled"
	ReminderStatusUnprocessable ReminderStatus = "unprocessable"
	ReminderStatusSendRequested ReminderStatus = "send_requested"
	ReminderStatusSent          ReminderStatus = "sent"
	ReminderStatusFailed        ReminderStatus = "failed"
	ReminderStatusDeleted       ReminderStatus = "deleted"
)

type AppointmentReminder struct {
	Status          ReminderStatus
	RemindBefore    time.Duration
	ScheduledAt     *time.Time
	SentRequestedAt *time.Time
	SentAt          *time.Time
	FailedAt        *time.Time
	FailureReason   *string
	UpdatedAt       time.Time
}

func NewAppointmentReminder(remindBefore time.Duration, now time.Time) (AppointmentReminder, error) {
	if remindBefore <= 0 {
		return AppointmentReminder{}, ErrInvalidReminder
	}
	return AppointmentReminder{
		Status:       ReminderStatusPending,
		RemindBefore: remindBefore,
		UpdatedAt:    now.UTC(),
	}, nil
}

func ReconstituteAppointmentReminder(reminder AppointmentReminder) (AppointmentReminder, error) {
	if !reminder.Status.Valid() || reminder.RemindBefore <= 0 {
		return AppointmentReminder{}, ErrInvalidReminder
	}
	reminder.ScheduledAt = utcTimePointer(reminder.ScheduledAt)
	reminder.SentRequestedAt = utcTimePointer(reminder.SentRequestedAt)
	reminder.SentAt = utcTimePointer(reminder.SentAt)
	reminder.FailedAt = utcTimePointer(reminder.FailedAt)
	reminder.UpdatedAt = reminder.UpdatedAt.UTC()
	return reminder, nil
}

func (status ReminderStatus) Valid() bool {
	switch status {
	case ReminderStatusPending,
		ReminderStatusScheduled,
		ReminderStatusUnprocessable,
		ReminderStatusSendRequested,
		ReminderStatusSent,
		ReminderStatusFailed,
		ReminderStatusDeleted:
		return true
	default:
		return false
	}
}

func (reminder *AppointmentReminder) Schedule(sendAt time.Time, now time.Time) error {
	if !sendAt.After(now.UTC()) {
		return ErrInvalidReminder
	}
	scheduledAt := sendAt.UTC()
	reminder.Status = ReminderStatusScheduled
	reminder.ScheduledAt = &scheduledAt
	reminder.UpdatedAt = now.UTC()
	return nil
}

func (reminder *AppointmentReminder) MarkUnprocessable(reason string, now time.Time) {
	reminder.Status = ReminderStatusUnprocessable
	reminder.FailureReason = optionalString(reason)
	reminder.UpdatedAt = now.UTC()
}

func (reminder *AppointmentReminder) MarkSendRequested(now time.Time) {
	sentRequestedAt := now.UTC()
	reminder.Status = ReminderStatusSendRequested
	reminder.SentRequestedAt = &sentRequestedAt
	reminder.UpdatedAt = sentRequestedAt
}

func (reminder *AppointmentReminder) MarkSent(now time.Time) {
	sentAt := now.UTC()
	reminder.Status = ReminderStatusSent
	reminder.SentAt = &sentAt
	reminder.UpdatedAt = sentAt
}

func (reminder *AppointmentReminder) MarkFailed(reason string, now time.Time) {
	failedAt := now.UTC()
	reminder.Status = ReminderStatusFailed
	reminder.FailedAt = &failedAt
	reminder.FailureReason = optionalString(reason)
	reminder.UpdatedAt = failedAt
}

func (reminder *AppointmentReminder) MarkDeleted(now time.Time) {
	reminder.Status = ReminderStatusDeleted
	reminder.UpdatedAt = now.UTC()
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func utcTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
