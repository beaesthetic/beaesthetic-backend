package io.github.petretiandrea.scheduler.core.runtime.impl

import io.github.petretiandrea.scheduler.core.ScheduleJobRepository
import io.github.petretiandrea.scheduler.core.consumer.ConsumerStrategy
import java.time.Duration
import java.time.Instant
import java.util.concurrent.ScheduledFuture
import kotlin.time.toKotlinDuration
import kotlinx.coroutines.runBlocking
import org.springframework.scheduling.TaskScheduler
import org.springframework.scheduling.support.PeriodicTrigger

class SpringSchedulerRuntime(
    private val pollingInterval: Duration,
    jobRepository: ScheduleJobRepository,
    consumerStrategy: ConsumerStrategy,
    private val taskScheduler: TaskScheduler,
) : AbstractRuntime(jobRepository, consumerStrategy) {

    private var schedule: ScheduledFuture<*>? = null

    override suspend fun start() {
        val periodicTrigger = PeriodicTrigger(pollingInterval)
        if (schedule == null) {
            schedule =
                taskScheduler.schedule(
                    { runBlocking { tick(Instant.now(), pollingInterval.toKotlinDuration()) } },
                    periodicTrigger,
                )
        }
    }

    override suspend fun stop() {
        schedule?.cancel(true)
        schedule = null
    }
}
