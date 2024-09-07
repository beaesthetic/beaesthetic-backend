package it.beaesthetic.appointment.agenda.application.events

import arrow.core.flatMap
import it.beaesthetic.appointment.agenda.domain.event.*
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

    suspend fun handle(command: EditAgendaSchedule): Result<AgendaEvent> =
        kotlin
            .runCatching {
                val (schedule, version) =
                    agendaRepository.findEvent(command.scheduleId)
                        ?: throw IllegalArgumentException("Schedule not found")

                val scheduleData =
                    when (val data = schedule.data) {
                        is BasicEventData ->
                            data.copy(
                                title = command.title ?: data.title,
                                description = command.description ?: data.description
                            )
                        is AppointmentEventData ->
                            data.copy(services = command.services?.toList() ?: data.services)
                    }
                val updatedSchedule =
                    if (command.timeSpan != null) schedule.reschedule(command.timeSpan)
                    else schedule

                updatedSchedule.copy(data = scheduleData).versioned(version)
            }
            .flatMap { agendaRepository.saveEvent(it.entity, expectedVersion = it.version) }
}
