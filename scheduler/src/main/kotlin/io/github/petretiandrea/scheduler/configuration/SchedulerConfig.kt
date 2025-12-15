package io.github.petretiandrea.scheduler.configuration

import java.time.Duration
import org.springframework.boot.context.properties.ConfigurationProperties

@ConfigurationProperties(prefix = "scheduler")
data class SchedulerConfig(
    val name: String,
    val pollingInterval: Duration,
    val peekLeaseTtl: Duration,
    val peekBatchSize: Int,
)
