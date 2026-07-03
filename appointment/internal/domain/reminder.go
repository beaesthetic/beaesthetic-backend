package domain

type ReminderStatus string

const (
	ReminderPending       ReminderStatus = "PENDING"
	ReminderScheduled     ReminderStatus = "SCHEDULED"
	ReminderSentRequested ReminderStatus = "SENT_REQUESTED"
	ReminderSent          ReminderStatus = "SENT"
	ReminderDeleted       ReminderStatus = "DELETED"
	ReminderFailToSend    ReminderStatus = "FAIL_TO_SEND"
	ReminderUnprocessable ReminderStatus = "UNPROCESSABLE"
)
