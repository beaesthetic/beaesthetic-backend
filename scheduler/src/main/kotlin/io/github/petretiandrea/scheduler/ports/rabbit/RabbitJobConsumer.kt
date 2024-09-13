package io.github.petretiandrea.scheduler.ports.rabbit

import io.github.petretiandrea.scheduler.core.ScheduleJob
import io.github.petretiandrea.scheduler.core.consumer.JobConsumer
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.slf4j.LoggerFactory
import org.springframework.amqp.core.MessageProperties
import org.springframework.amqp.rabbit.core.RabbitTemplate
import org.springframework.amqp.support.converter.Jackson2JsonMessageConverter
import org.springframework.stereotype.Component

@Component
class RabbitJobConsumer(private val rabbitTemplate: RabbitTemplate) : JobConsumer {

    private val jackson2JsonMessageConverter = Jackson2JsonMessageConverter()
    private val log = LoggerFactory.getLogger(RabbitJobConsumer::class.java)

    override suspend fun consume(job: ScheduleJob) =
        withContext(Dispatchers.IO) {
            log.info("Publishing schedule job to Rabbit ${job.id} to route ${job.meta.route}")
            val message = jackson2JsonMessageConverter.toMessage(job.meta.data, MessageProperties())

            rabbitTemplate.send("", job.meta.route, message)
        }
}
