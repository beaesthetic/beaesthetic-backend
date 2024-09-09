package it.beaesthetic.appointment.agenda.application.events

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.event.TimeSpan
import jakarta.enterprise.context.ApplicationScoped

@ApplicationScoped
class AgendaQueryHandler(private val repository: AgendaRepository) {

    suspend fun handle(activityId: AgendaEventId): Result<AgendaEvent> {
        val (schedule, _) =
            repository.findEvent(activityId)
                ?: return Result.failure(Throwable("No schedule found for activity id $activityId"))
        return Result.success(schedule)
    }

    suspend fun handle(timeSpan: TimeSpan): List<AgendaEvent> {
        return repository.findEvents(timeSpan)
    }

    suspend fun handleByAttendeeId(attendeeId: String): List<AgendaEvent> {
        return repository.findByAttendeeId(attendeeId)
    }
}
