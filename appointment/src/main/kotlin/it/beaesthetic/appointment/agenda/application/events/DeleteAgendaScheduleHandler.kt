package it.beaesthetic.appointment.agenda.application.events

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.event.CancelReason
import jakarta.enterprise.context.ApplicationScoped

data class DeleteAgendaSchedule(val id: AgendaEventId, val reason: CancelReason)

@ApplicationScoped
class DeleteAgendaScheduleHandler(private val agendaRepository: AgendaRepository) {

    suspend fun handle(command: DeleteAgendaSchedule): Result<AgendaEvent> {
        val (schedule, version) =
            agendaRepository.findEvent(command.id)
                ?: throw IllegalArgumentException("Schedule not found")

        return agendaRepository.saveEvent(schedule.cancel(command.reason), version)
    }
}
