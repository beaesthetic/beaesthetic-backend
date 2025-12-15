package io.github.petretiandrea.scheduler.core

import io.github.petretiandrea.scheduler.core.consumer.ConsumerStrategy
import io.github.petretiandrea.scheduler.core.runtime.SchedulerRuntime
import io.github.petretiandrea.scheduler.core.runtime.impl.DefaultRuntime
import io.github.petretiandrea.scheduler.core.runtime.impl.SpringSchedulerRuntime
import java.time.Duration
import kotlinx.coroutines.CoroutineScope
import org.springframework.scheduling.TaskScheduler

class SchedulerFactory(
    private val schedulerName: String,
    private val jobRepository: ScheduleJobRepository,
    private val pollingInterval: Duration,
    private val strategy: ConsumerStrategy,
) {

    fun createScheduler(): Scheduler {
        return BasicScheduler(schedulerName, jobRepository)
    }

    fun createDefaultRuntime(coroutineScope: CoroutineScope): SchedulerRuntime {
        return DefaultRuntime(
            coroutineScope = coroutineScope,
            pollingInterval = pollingInterval,
            jobRepository = jobRepository,
            consumerStrategy = strategy,
        )
    }

    fun createSpringRuntime(taskScheduler: TaskScheduler): SchedulerRuntime {
        return SpringSchedulerRuntime(
            jobRepository = jobRepository,
            consumerStrategy = strategy,
            taskScheduler = taskScheduler,
            pollingInterval = pollingInterval,
        )
    }

    companion object {
        fun builder(name: String, jobRepository: ScheduleJobRepository): Builder {
            return Builder(name, jobRepository)
        }
    }

    class Builder(private val name: String, private val jobRepository: ScheduleJobRepository) {
        companion object {
            private val DEFAULT_POLLING_INTERVAL: Duration = Duration.ofSeconds(10)
            private val DEFAULT_CONSUMER_STRATEGY: ConsumerStrategy = ConsumerStrategy.atMostOnce()
        }

        private var pollingInterval: Duration? = null
        private var strategy: ConsumerStrategy? = null

        fun pollingInterval(pollingInterval: Duration): Builder {
            this.pollingInterval = pollingInterval
            return this
        }

        fun strategy(strategy: ConsumerStrategy): Builder {
            this.strategy = strategy
            return this
        }

        fun build(): SchedulerFactory {
            return SchedulerFactory(
                name,
                jobRepository,
                pollingInterval ?: DEFAULT_POLLING_INTERVAL,
                strategy ?: DEFAULT_CONSUMER_STRATEGY,
            )
        }
    }
}
