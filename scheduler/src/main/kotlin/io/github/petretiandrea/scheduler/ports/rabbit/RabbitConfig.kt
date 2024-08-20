package io.github.petretiandrea.scheduler.ports.rabbit

import org.springframework.amqp.core.Queue
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration

@Configuration
class RabbitConfig {

    @Bean
    fun createJobQueue(): Queue {
        return Queue("jobs")
    }
}
