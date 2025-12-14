package it.beaesthetic.appointment.agenda.application.reminder

import arrow.core.flatMap
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderService
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import jakarta.enterprise.context.ApplicationScoped
import org.jboss.logging.Logger

data class ScheduleReminder(val eventId: AgendaEventId)

@ApplicationScoped
class ScheduleReminderHandler(
    private val agendaRepository: AgendaRepository,
    private val reminderService: ReminderService,
) {

    private val log = Logger.getLogger(ScheduleReminderHandler::class.java)

    suspend fun handle(command: ScheduleReminder): Result<Unit> {
        log.info("Scheduling reminder for event ${command.eventId}")
        val (event, version) =
            agendaRepository.findEvent(command.eventId)
                ?: return Result.failure(
                    IllegalArgumentException("Event with id ${command.eventId} not found")
                )

        return reminderService
            .scheduleReminder(event)
            .onSuccess {
                if (it.reminder.status != ReminderStatus.SCHEDULED) {
                    log.info("Not scheduling reminder in state ${it.reminder.status}")
                }
            }
            .onFailure { log.error("Failed to schedule reminder for ${event.id}", it) }
            .flatMap { agendaRepository.saveEvent(it, version) }
            .map {}
    }
}
