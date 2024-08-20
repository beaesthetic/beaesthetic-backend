package io.github.petretiandrea.scheduler.core.consumer

import io.github.petretiandrea.scheduler.core.Job

interface ConsumerStrategy {
    suspend fun consumeWithAck(job: Job, consumers: List<JobConsumer>)

    companion object {
        fun atMostOnce(): ConsumerStrategy = AtMostOnceStrategy()
        fun atLeastOnce(): ConsumerStrategy = AtLeastOnceStrategy()
    }
}
