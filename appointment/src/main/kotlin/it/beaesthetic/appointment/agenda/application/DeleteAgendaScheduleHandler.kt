package it.beaesthetic.appointment.agenda.application

import it.beaesthetic.appointment.agenda.domain.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.AgendaSchedule
import it.beaesthetic.appointment.agenda.domain.CancelReason
import jakarta.enterprise.context.ApplicationScoped
import java.util.UUID

data class DeleteAgendaSchedule(val id: UUID, val reason: CancelReason)

@ApplicationScoped
class DeleteAgendaScheduleHandler(private val agendaRepository: AgendaRepository) {

    suspend fun handle(command: DeleteAgendaSchedule): Result<AgendaSchedule> {
        val (schedule, version) =
            agendaRepository.findSchedule(command.id.toString())
                ?: throw IllegalArgumentException("Schedule not found")

        return agendaRepository.saveSchedule(schedule.cancel(command.reason), version)
    }
}
