package it.beaesthetic.appointment.agenda.application.reminder

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import jakarta.enterprise.context.ApplicationScoped
import org.jboss.logging.Logger

data class ConfirmReminderSent(val eventId: String)

@ApplicationScoped
class ConfirmReminderSentHandler(
    private val agendaRepository: AgendaRepository,
) {

    private val log = Logger.getLogger(SendAgendaScheduleReminderHandler::class.java)

    suspend fun handle(command: ConfirmReminderSent): Result<AgendaEvent> {
        log.info("Confirm sent reminder for ${command.eventId}")
        val (event, version) =
            agendaRepository.findEvent(command.eventId)
                ?: throw IllegalArgumentException("No event with id ${command.eventId} found")

        return agendaRepository.saveEvent(event.updateReminderStatus(ReminderStatus.SENT), version)
    }
}
