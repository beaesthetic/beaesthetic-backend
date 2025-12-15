package it.beaesthetic.notification

import io.quarkus.runtime.Startup
import io.smallrye.mutiny.coroutines.awaitSuspending
import io.vertx.core.json.JsonObject
import io.vertx.mutiny.core.Vertx
import io.vertx.mutiny.rabbitmq.RabbitMQClient
import io.vertx.rabbitmq.RabbitMQOptions
import jakarta.annotation.Priority
import jakarta.enterprise.context.ApplicationScoped
import kotlinx.coroutines.runBlocking
import org.eclipse.microprofile.config.Config
import org.eclipse.microprofile.config.inject.ConfigProperty
import org.jboss.logging.Logger

@ApplicationScoped
class SetupRabbit(private val config: Config) {

    private val log = Logger.getLogger(SetupRabbit::class.java)

    @Startup
    @Priority(1)
    fun setupRetryQueue(@ConfigProperty(name = "queues.notification") queueName: String) =
        runBlocking {
            val vertx = Vertx.vertx()
            try {
                val rabbitClient =
                    RabbitMQClient.create(
                        vertx,
                        RabbitMQOptions().apply {
                            host = config.getValue("rabbitmq-host", String::class.java)
                            port = config.getValue("rabbitmq-port", Int::class.java)
                            user = config.getValue("rabbitmq-username", String::class.java)
                            password = config.getValue("rabbitmq-password", String::class.java)
                        },
                    )
                val retryQueue = "${queueName}-retry"
                val retryExchange = "$retryQueue-exchange"
                rabbitClient
                    .start()
                    .onItem()
                    .invoke { _ -> log.info("Setting up retry queue for $queueName") }
                    .flatMap { rabbitClient.exchangeDeclare(retryExchange, "fanout", true, false) }
                    .flatMap {
                        rabbitClient.queueDeclare(
                            retryQueue,
                            true,
                            false,
                            false,
                            JsonObject.of(
                                "x-dead-letter-exchange",
                                "amq.direct",
                                "x-message-ttl",
                                1,
                            ),
                        )
                    }
                    .flatMap {
                        rabbitClient.queueDeclare(
                            queueName,
                            true,
                            false,
                            false,
                            JsonObject.of("x-dead-letter-exchange", retryQueue),
                        )
                    }
                    .flatMap { rabbitClient.queueBind(queueName, "amq.direct", queueName) }
                    .flatMap { rabbitClient.queueBind(retryQueue, retryExchange, retryQueue) }
                    .onItem()
                    .invoke { _ -> log.info("Setup rabbit done for $queueName") }
                    .onFailure()
                    .invoke { e -> log.error("Failed to setup queue $queueName", e) }
                    .awaitSuspending()
            } finally {
                vertx.close().awaitSuspending()
            }
        }
}
