package it.beaesthetic.appointment.agenda.application.events

import arrow.core.flatMap
import it.beaesthetic.appointment.agenda.domain.event.*
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderOptions
import jakarta.enterprise.context.ApplicationScoped
import java.time.Duration
import java.util.*

data class CreateAgendaSchedule(
    val timeSpan: TimeSpan,
    val data: AgendaEventData,
    val attendeeId: String
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
                    when (command.data) {
                        //                        is AppointmentEventData ->
                        //
                        // customerRegistry.findByCustomerId(command.attendeeId)?.let {
                        //                                Attendee(it.customerId, it.displayName)
                        //                            }
                        else -> Attendee(command.attendeeId, "self")
                    }
                        ?: throw IllegalArgumentException("Unknown attendee ${command.attendeeId}")

                AgendaEvent.create(
                    id = AgendaEventId(UUID.randomUUID().toString()),
                    timeSpan = command.timeSpan,
                    attendee = attendee,
                    data = command.data,
                    reminderOptions = ReminderOptions(triggerBefore = Duration.ofSeconds(10))
                )
            }
            .flatMap { agendaRepository.saveEvent(it) }
}
