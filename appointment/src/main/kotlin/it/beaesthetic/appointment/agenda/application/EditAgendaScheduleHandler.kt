package it.beaesthetic.appointment.agenda.application

import arrow.core.flatMap
import it.beaesthetic.appointment.agenda.domain.*
import it.beaesthetic.appointment.common.OptimisticConcurrency.versioned
import jakarta.enterprise.context.ApplicationScoped

data class EditAgendaSchedule(
    val scheduleId: String,
    val timeSpan: TimeSpan?,
    val services: Set<AppointmentService>?,
    val title: String?,
    val description: String?
)

@ApplicationScoped
class EditAgendaScheduleHandler(private val agendaRepository: AgendaRepository) {

    suspend fun handle(command: EditAgendaSchedule): Result<AgendaSchedule> =
        kotlin
            .runCatching {
                val (schedule, version) =
                    agendaRepository.findSchedule(command.scheduleId)
                        ?: throw IllegalArgumentException("Schedule not found")

                val scheduleData =
                    when (val data = schedule.data) {
                        is BasicScheduleData ->
                            data.copy(
                                title = command.title ?: data.title,
                                description = command.description ?: data.description
                            )
                        is AppointmentScheduleData ->
                            data.copy(services = command.services?.toList() ?: data.services)
                    }
                val updatedSchedule =
                    if (command.timeSpan != null) schedule.reschedule(command.timeSpan)
                    else schedule

                updatedSchedule.copy(data = scheduleData).versioned(version)
            }
            .flatMap { agendaRepository.saveSchedule(it.entity, expectedVersion = it.version) }
}
