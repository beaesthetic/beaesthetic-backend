package it.beaesthetic.appointment.agenda.domain.reminder

import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import java.time.Instant

data class Reminder(val eventId: AgendaEventId, val status: ReminderStatus, val sentAt: Instant?)
