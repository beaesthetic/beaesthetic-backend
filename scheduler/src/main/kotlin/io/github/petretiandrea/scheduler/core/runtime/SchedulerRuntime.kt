package io.github.petretiandrea.scheduler.core.runtime

import io.github.petretiandrea.scheduler.core.consumer.JobConsumer
import java.time.Instant
import kotlin.time.Duration

interface SchedulerRuntime {
    fun addConsumer(jobConsumer: JobConsumer)
    fun removeConsumer(jobConsumer: JobConsumer)
    suspend fun start()
    suspend fun stop()
    suspend fun tick(tick: Instant, delta: Duration)
}
