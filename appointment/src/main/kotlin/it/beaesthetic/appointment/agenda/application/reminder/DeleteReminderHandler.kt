package it.beaesthetic.appointment.agenda.application.reminder

import arrow.core.flatMap
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderService
import jakarta.enterprise.context.ApplicationScoped
import kotlin.math.log
import org.jboss.logging.Logger

data class DeleteReminder(val eventId: AgendaEventId)

@ApplicationScoped
class DeleteReminderHandler(
    private val agendaRepository: AgendaRepository,
    private val reminderService: ReminderService,
) {

    private val log = Logger.getLogger(DeleteReminderHandler::class.java)

    suspend fun handle(command: DeleteReminder): Result<Unit> = runCatching {
        val (event, version) =
            agendaRepository.findEvent(command.eventId)
                ?: throw IllegalArgumentException("Event with id ${command.eventId} not found")

        reminderService
            .unscheduleReminder(event)
            .flatMap { agendaRepository.saveEvent(it, version) }
            .onSuccess { log.info("Successfully deleted reminder") }
            .onFailure {
                log.warn("Failed to delete reminder with status ${event.reminder.status}", it)
            }
    }
}
