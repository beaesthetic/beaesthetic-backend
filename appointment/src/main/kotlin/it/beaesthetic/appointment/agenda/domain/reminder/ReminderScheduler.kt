package it.beaesthetic.appointment.agenda.domain.reminder

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent

interface ReminderScheduler {
    suspend fun scheduleReminder(event: AgendaEvent): AgendaEvent
    suspend fun unschedule(event: AgendaEvent): AgendaEvent
}
