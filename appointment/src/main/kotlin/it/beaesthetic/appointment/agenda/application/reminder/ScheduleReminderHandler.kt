package it.beaesthetic.appointment.agenda.application.reminder

import arrow.core.flatMap
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderScheduler
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import jakarta.enterprise.context.ApplicationScoped
import org.jboss.logging.Logger

data class ScheduleReminder(val eventId: AgendaEventId)

@ApplicationScoped
class ScheduleReminderHandler(
    private val agendaRepository: AgendaRepository,
    private val reminderScheduler: ReminderScheduler
) {

    private val log = Logger.getLogger(ScheduleReminderHandler::class.java)

    suspend fun handle(command: ScheduleReminder): Result<Unit> = runCatching {
        log.info("Scheduling reminder for event ${command.eventId}")
        val (event, version) =
            agendaRepository.findEvent(command.eventId)
                ?: throw IllegalArgumentException("Event with id ${command.eventId} not found")

        runCatching { reminderScheduler.scheduleReminder(event) }
            .onFailure { log.error("Failed to schedule reminder for ${event.id}", it) }
            .flatMap {
                agendaRepository.saveEvent(
                    event.updateReminderStatus(ReminderStatus.SCHEDULED),
                    version
                )
            }
            .getOrThrow()
    }
}
