package it.beaesthetic.appointment.agenda.application.reminder

import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderScheduler
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import jakarta.enterprise.context.ApplicationScoped

data class DeleteReminder(val eventId: AgendaEventId)

@ApplicationScoped
class DeleteReminderHandler(
    private val agendaRepository: AgendaRepository,
    private val reminderScheduler: ReminderScheduler
) {
    suspend fun handle(command: DeleteReminder): Result<Unit> = runCatching {
        val (event, version) =
            agendaRepository.findEvent(command.eventId)
                ?: throw IllegalArgumentException("Event with id ${command.eventId} not found")

        // TODO: remove reminder
        reminderScheduler.unschedule(event)

        agendaRepository.saveEvent(event.updateReminderStatus(ReminderStatus.DELETED), version)
    }
}
