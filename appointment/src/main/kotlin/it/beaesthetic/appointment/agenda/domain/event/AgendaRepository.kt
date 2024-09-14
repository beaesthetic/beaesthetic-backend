package it.beaesthetic.appointment.agenda.domain.event

import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import it.beaesthetic.appointment.common.OptimisticConcurrency

interface AgendaRepository {
    suspend fun findEvent(
        scheduleId: AgendaEventId
    ): OptimisticConcurrency.VersionedEntity<AgendaEvent>?
    suspend fun saveEvent(schedule: AgendaEvent, expectedVersion: Long = 0): Result<AgendaEvent>
    suspend fun findEvents(timeSpan: TimeSpan): List<AgendaEvent>
    suspend fun findByAttendeeId(attendeeId: String): List<AgendaEvent>

    suspend fun findEventsWithReminderState(reminderStatus: ReminderStatus): List<AgendaEvent>
    suspend fun deleteEvent(scheduleId: AgendaEventId): Result<Boolean>
}
