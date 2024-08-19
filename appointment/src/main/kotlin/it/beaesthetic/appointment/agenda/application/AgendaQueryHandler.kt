package it.beaesthetic.appointment.agenda.application

import it.beaesthetic.appointment.agenda.domain.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.AgendaSchedule
import it.beaesthetic.appointment.agenda.domain.TimeSpan
import jakarta.enterprise.context.ApplicationScoped

@ApplicationScoped
class AgendaQueryHandler(private val repository: AgendaRepository) {

    suspend fun handle(activityId: String): Result<AgendaSchedule> {
        val (schedule, _) =
            repository.findSchedule(activityId)
                ?: return Result.failure(Throwable("No schedule found for activity id $activityId"))
        return Result.success(schedule)
    }

    suspend fun handle(timeSpan: TimeSpan): List<AgendaSchedule> {
        return repository.findSchedules(timeSpan)
    }

    suspend fun handleByAttendeeId(attendeeId: String): List<AgendaSchedule> {
        return repository.findByAttendeeId(attendeeId)
    }
}
