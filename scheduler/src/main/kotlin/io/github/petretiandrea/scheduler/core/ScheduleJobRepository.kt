package io.github.petretiandrea.scheduler.core

import java.time.Instant
import kotlin.time.Duration

interface ScheduleJobRepository {
    suspend fun save(job: ScheduleJob): Result<ScheduleJob>
    suspend fun findById(id: ScheduleId): Result<ScheduleJob>
    suspend fun pollJobs(tick: Instant, delta: Duration): List<Job>
    suspend fun remove(scheduleId: ScheduleId): Result<Unit>
}

interface Job {
    val scheduleJob: ScheduleJob
    suspend fun ack()
    suspend fun nack(reason: String?)
}
