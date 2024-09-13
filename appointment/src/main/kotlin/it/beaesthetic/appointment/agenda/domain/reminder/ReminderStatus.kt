package it.beaesthetic.appointment.agenda.domain.reminder

enum class ReminderStatus {
    PENDING,
    SCHEDULED,
    SENT_REQUESTED,
    SENT,
    FAIL_TO_SEND,
    UNPROCESSABLE,
    DELETED
}
