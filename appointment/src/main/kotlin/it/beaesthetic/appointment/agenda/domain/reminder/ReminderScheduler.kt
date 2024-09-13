package it.beaesthetic.appointment.agenda.domain.reminder

import java.time.Instant

/** A reminder scheduler allows to schedule a reminder for given Agenda Event */
interface ReminderScheduler {
    suspend fun scheduleReminder(reminder: Reminder, sendAt: Instant): Reminder
    suspend fun unschedule(reminder: Reminder): Reminder
}
