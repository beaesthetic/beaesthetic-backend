package io.github.petretiandrea.scheduler.core

import java.time.Instant
import java.util.*

class BasicScheduler(
    override val name: String,
    private val scheduleJobRepository: ScheduleJobRepository
) : Scheduler {

    override suspend fun schedule(scheduleAt: Instant, meta: ScheduleMeta): Result<ScheduleJob> {
        return schedule(UUID.randomUUID().toString(), scheduleAt, meta)
    }

    override suspend fun schedule(
        scheduleId: String,
        scheduleAt: Instant,
        meta: ScheduleMeta
    ): Result<ScheduleJob> {
        val schedule =
            ScheduleJob(id = ScheduleId(scheduleId), meta = meta, scheduleAt = scheduleAt)

        return scheduleJobRepository.save(schedule)
    }

    override suspend fun delete(id: ScheduleId): Result<Unit> {
        return scheduleJobRepository.remove(id)
    }
}
