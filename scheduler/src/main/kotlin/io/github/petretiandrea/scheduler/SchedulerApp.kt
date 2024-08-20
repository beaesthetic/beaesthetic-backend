package io.github.petretiandrea.scheduler

import io.github.petretiandrea.scheduler.configuration.SchedulerConfig
import io.github.petretiandrea.scheduler.core.consumer.JobConsumer
import io.github.petretiandrea.scheduler.core.runtime.SchedulerRuntime
import kotlinx.coroutines.runBlocking
import org.slf4j.LoggerFactory
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.context.event.ApplicationReadyEvent
import org.springframework.boot.context.properties.EnableConfigurationProperties
import org.springframework.boot.runApplication
import org.springframework.context.ApplicationEvent
import org.springframework.context.ApplicationListener
import org.springframework.context.event.ContextClosedEvent
import org.springframework.scheduling.annotation.EnableScheduling

@SpringBootApplication
@EnableScheduling
@EnableConfigurationProperties(value = [SchedulerConfig::class])
class SchedulerApp(private val schedulerRuntime: SchedulerRuntime) :
    ApplicationListener<ApplicationEvent> {

    private val logger = LoggerFactory.getLogger(SchedulerApp::class.java)

    override fun onApplicationEvent(event: ApplicationEvent) = runBlocking {
        when (event) {
            is ApplicationReadyEvent -> {
                logger.info("Registering scheduler consumers")
                val consumers =
                    event.applicationContext.getBeansOfType(JobConsumer::class.java).values
                logger.info("Found ${consumers.size} consumers")
                consumers.forEach { schedulerRuntime.addConsumer(it) }
                schedulerRuntime.start()
                logger.info("Started scheduler runtime")
            }
            is ContextClosedEvent -> {
                logger.info("Stopping scheduler runtime")
                schedulerRuntime.stop()
                logger.info("Stopped scheduler runtime")
            }
        }
    }
}

fun main() {
    runApplication<SchedulerApp>()
}
