package io.github.petretiandrea.scheduler.core.runtime.impl

import io.github.petretiandrea.scheduler.core.ScheduleJobRepository
import io.github.petretiandrea.scheduler.core.consumer.ConsumerStrategy
import io.github.petretiandrea.scheduler.core.consumer.JobConsumer
import io.github.petretiandrea.scheduler.core.runtime.SchedulerRuntime
import java.time.Instant
import kotlin.time.Duration
import org.slf4j.LoggerFactory

abstract class AbstractRuntime(
    private val jobRepository: ScheduleJobRepository,
    private val consumerStrategy: ConsumerStrategy,
) : SchedulerRuntime {

    private var consumers = emptyList<JobConsumer>()
    private val log = LoggerFactory.getLogger(SchedulerRuntime::class.java)

    override fun addConsumer(jobConsumer: JobConsumer) {
        consumers += jobConsumer
    }

    override fun removeConsumer(jobConsumer: JobConsumer) {
        consumers -= jobConsumer
    }

    override suspend fun tick(tick: Instant, delta: Duration) {
        val jobs = jobRepository.pollJobs(tick, delta)
        if (jobs.isNotEmpty()) {
            log.info("Polled ${jobs.size} jobs at $tick")
        }
        jobs.forEach { consumerStrategy.consumeWithAck(it, consumers) }
    }
}
