package io.github.petretiandrea.scheduler.core.consumer

import io.github.petretiandrea.scheduler.core.Job

class AtLeastOnceStrategy : ConsumerStrategy {
    override suspend fun consumeWithAck(job: Job, consumers: List<JobConsumer>) {
        kotlin
            .runCatching { consumers.forEach { it.consume(job.scheduleJob) } }
            .onSuccess { job.ack() }
            .onFailure { job.nack(it.message) }
    }
}
