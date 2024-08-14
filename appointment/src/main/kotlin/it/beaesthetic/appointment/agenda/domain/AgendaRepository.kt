package it.beaesthetic.appointment.agenda.domain

import it.beaesthetic.appointment.common.OptimisticConcurrency

interface AgendaRepository {
    suspend fun findSchedule(scheduleId: String): OptimisticConcurrency.VersionedEntity<AgendaSchedule>?
    suspend fun saveSchedule(schedule: AgendaSchedule, expectedVersion: Long = 0): Result<AgendaSchedule>
    suspend fun findSchedules(timeSpan: TimeSpan): List<AgendaSchedule>
    suspend fun findByAttendeeId(attendeeId: String): List<AgendaSchedule>

    suspend fun deleteSchedule(scheduleId: String): Result<Boolean>
}