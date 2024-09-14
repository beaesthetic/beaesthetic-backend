package io.github.petretiandrea.scheduler

import com.fasterxml.jackson.databind.ObjectMapper
import io.github.petretiandrea.scheduler.configuration.SchedulerConfig
import io.github.petretiandrea.scheduler.core.*
import io.github.petretiandrea.scheduler.core.consumer.ConsumerStrategy
import io.github.petretiandrea.scheduler.core.runtime.SchedulerRuntime
import io.github.petretiandrea.scheduler.redis.RedisJobRepository
import io.github.petretiandrea.scheduler.redis.ScheduleJobRedisOptions
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.context.annotation.ImportRuntimeHints
import org.springframework.data.redis.core.ReactiveRedisTemplate
import org.springframework.data.redis.core.RedisTemplate
import org.springframework.scheduling.TaskScheduler

@Configuration
@ImportRuntimeHints(RedisJobRepository.RedisJobStoreRuntimeHints::class)
class SchedulerDI {

    @Bean
    fun schedulerFactory(
        scheduleJobRepository: RedisJobRepository,
        schedulerConfig: SchedulerConfig,
    ): SchedulerFactory {
        return SchedulerFactory.builder(schedulerConfig.name, scheduleJobRepository)
            .strategy(ConsumerStrategy.atLeastOnce())
            .pollingInterval(schedulerConfig.pollingInterval)
            .build()
    }

    @Bean
    fun scheduler(schedulerFactory: SchedulerFactory): Scheduler {
        return schedulerFactory.createScheduler()
    }

    @Bean
    fun schedulerRuntime(
        schedulerFactory: SchedulerFactory,
        taskScheduler: TaskScheduler,
    ): SchedulerRuntime {
        return schedulerFactory.createSpringRuntime(taskScheduler)
    }

    @Bean
    fun scheduleJobRepository(
        schedulerConfig: SchedulerConfig,
        objectMapper: ObjectMapper,
        reactiveRedisTemplate: ReactiveRedisTemplate<String, String>,
        redisTemplate: RedisTemplate<String, String>,
    ): RedisJobRepository {
        return RedisJobRepository(
            reactiveRedisTemplate = reactiveRedisTemplate,
            redisTemplate = redisTemplate,
            objectMapper = objectMapper,
            scheduleJobRedisOptions =
                ScheduleJobRedisOptions(
                    sortedSetName = "${schedulerConfig.name}-clock",
                    taskSetName = "${schedulerConfig.name}-tasks",
                    peekBatchSize = schedulerConfig.peekBatchSize,
                    peekLeaseTTL = schedulerConfig.peekLeaseTtl
                )
        )
    }
}
