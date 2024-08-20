package io.github.petretiandrea.scheduler.core.consumer

import io.github.petretiandrea.scheduler.core.Job

class AtMostOnceStrategy : ConsumerStrategy {
    override suspend fun consumeWithAck(job: Job, consumers: List<JobConsumer>) {
        job.ack()
        consumers.forEach { it.consume(job.scheduleJob) }
    }
}
