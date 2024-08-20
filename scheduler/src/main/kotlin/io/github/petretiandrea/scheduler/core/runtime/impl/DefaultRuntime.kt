package io.github.petretiandrea.scheduler.core.runtime.impl

import io.github.petretiandrea.scheduler.core.ScheduleJobRepository
import io.github.petretiandrea.scheduler.core.consumer.ConsumerStrategy
import java.time.Duration
import java.time.Instant
import kotlin.time.toKotlinDuration
import kotlinx.coroutines.*

class DefaultRuntime(
    private val coroutineScope: CoroutineScope,
    private val pollingInterval: Duration,
    jobRepository: ScheduleJobRepository,
    consumerStrategy: ConsumerStrategy
) : AbstractRuntime(jobRepository, consumerStrategy) {

    private var job: Job? = null

    override suspend fun start() {
        if (job == null) {
            job =
                coroutineScope.launch {
                    tick(Instant.now(), pollingInterval.toKotlinDuration())
                    delay(pollingInterval.toKotlinDuration())
                }
        }
    }

    override suspend fun stop() {
        job?.cancelAndJoin()
        job = null
    }
}
