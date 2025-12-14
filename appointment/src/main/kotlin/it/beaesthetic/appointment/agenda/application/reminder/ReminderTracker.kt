package it.beaesthetic.appointment.agenda.application.reminder

import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus

interface ReminderTracker {
    fun trackReminderState(reminderStatus: ReminderStatus)

    fun trackFailedReminders(failedReminder: Int)
}
