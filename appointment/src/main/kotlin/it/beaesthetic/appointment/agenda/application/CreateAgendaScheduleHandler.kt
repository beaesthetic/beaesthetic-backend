package it.beaesthetic.appointment.agenda.application

import arrow.core.flatMap
import it.beaesthetic.appointment.agenda.domain.*
import jakarta.enterprise.context.ApplicationScoped
import java.time.Instant
import java.util.*

data class CreateAgendaSchedule(
    val timeSpan: TimeSpan,
    val data: AgendaScheduleData,
    val attendeeId: String
)

@ApplicationScoped
class CreateAgendaScheduleHandler(
    private val agendaRepository: AgendaRepository,
    private val customerRegistry: CustomerRegistry
) {

    suspend fun handle(command: CreateAgendaSchedule): Result<AgendaSchedule> =
        kotlin
            .runCatching {
                val attendee =
                    when (command.data) {
                        is AppointmentScheduleData ->
                            customerRegistry.findByCustomerId(command.attendeeId)?.let {
                                Attendee(it.customerId, it.displayName)
                            }
                        is BasicScheduleData -> Attendee(command.attendeeId, "self")
                    }
                        ?: throw IllegalArgumentException("Unknown attendee ${command.attendeeId}")

                AgendaSchedule(
                    id = UUID.randomUUID().toString(),
                    timeSpan = command.timeSpan,
                    attendee = attendee,
                    createdAt = Instant.now(),
                    cancelReason = null,
                    data = command.data,
                )
            }
            .flatMap { agendaRepository.saveSchedule(it) }
}
