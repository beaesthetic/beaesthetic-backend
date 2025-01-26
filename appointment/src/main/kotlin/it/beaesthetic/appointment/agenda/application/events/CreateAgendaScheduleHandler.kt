package it.beaesthetic.appointment.agenda.application.events

import arrow.core.flatMap
import it.beaesthetic.appointment.agenda.domain.event.*
import jakarta.enterprise.context.ApplicationScoped
import java.time.Duration
import java.util.*

data class CreateAgendaSchedule(
    val timeSpan: TimeSpan,
    val data: AgendaEventData,
    val attendeeId: String,
    val triggerBefore: Duration
)

@ApplicationScoped
class CreateAgendaScheduleHandler(
    private val agendaRepository: AgendaRepository,
    private val customerRegistry: CustomerRegistry
) {

    suspend fun handle(command: CreateAgendaSchedule): Result<AgendaEvent> =
        kotlin
            .runCatching {
                val attendee =
                    //                    when (command.data) {
                    //                        is AppointmentEventData ->
                    //
                    // customerRegistry.findByCustomerId(command.attendeeId)?.let {
                    //                                Attendee(it.customerId, it.displayName)
                    //                            }
                    /*else -> */ Attendee(command.attendeeId, "self")
                //                    }
                //                        ?: throw IllegalArgumentException("Unknown attendee
                // ${command.attendeeId}")

                AgendaEvent.create(
                    id = AgendaEventId(UUID.randomUUID().toString()),
                    timeSpan = command.timeSpan,
                    attendee = attendee,
                    data = command.data,
                    reminderBefore = command.triggerBefore,
                )
            }
            .flatMap { agendaRepository.saveEvent(it) }
}
