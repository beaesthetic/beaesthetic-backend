package io.github.petretiandrea.scheduler.core.consumer

import io.github.petretiandrea.scheduler.core.ScheduleJob

interface JobConsumer {
    suspend fun consume(job: ScheduleJob)
}
