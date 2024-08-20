package io.github.petretiandrea.scheduler.core

import java.time.Instant

interface Scheduler {
    val name: String

    suspend fun schedule(scheduleAt: Instant, meta: ScheduleMeta): Result<ScheduleJob>

    suspend fun schedule(
        scheduleId: String,
        scheduleAt: Instant,
        meta: ScheduleMeta
    ): Result<ScheduleJob>

    suspend fun delete(id: ScheduleId): Result<Unit>
}
