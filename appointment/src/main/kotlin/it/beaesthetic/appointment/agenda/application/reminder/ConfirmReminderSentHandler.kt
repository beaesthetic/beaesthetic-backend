package it.beaesthetic.appointment.agenda.application.reminder

import it.beaesthetic.appointment.agenda.domain.Clock
import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import jakarta.enterprise.context.ApplicationScoped
import org.jboss.logging.Logger

data class ConfirmReminderSent(val eventId: AgendaEventId)

@ApplicationScoped
class ConfirmReminderSentHandler(
    private val agendaRepository: AgendaRepository,
    private val reminderTracker: ReminderTracker,
) {

    private val log = Logger.getLogger(SendAgendaScheduleReminderHandler::class.java)
    private val clock = Clock.default()

    suspend fun handle(command: ConfirmReminderSent): Result<AgendaEvent> {
        log.info("Confirm sent reminder for ${command.eventId}")
        val (event, version) =
            agendaRepository.findEvent(command.eventId)
                ?: throw IllegalArgumentException("No event with id ${command.eventId} found")

        return agendaRepository
            .saveEvent(event.trackReminderAsSent(clock = clock), version)
            .onSuccess { reminderTracker.trackReminderState(it.reminder.status) }
    }
}
