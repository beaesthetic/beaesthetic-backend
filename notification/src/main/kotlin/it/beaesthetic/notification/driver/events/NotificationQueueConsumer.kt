package it.beaesthetic.notification.driver.events

import io.vertx.core.json.JsonObject
import it.beaesthetic.notification.application.NotificationService
import it.beaesthetic.notification.domain.NotificationCreated
import it.beaesthetic.notification.domain.NotificationEvent
import it.beaesthetic.notification.domain.NotificationSent
import jakarta.enterprise.context.ApplicationScoped
import org.eclipse.microprofile.reactive.messaging.Incoming
import org.eclipse.microprofile.reactive.messaging.Message
import org.jboss.logging.Logger
import java.util.concurrent.CompletionStage

@ApplicationScoped
class NotificationQueueConsumer(
    private val notificationService: NotificationService
) {

    private val log = Logger.getLogger(NotificationQueueConsumer::class.java)

    @Incoming("notifications")
    suspend fun handle(message: Message<JsonObject>): CompletionStage<Void>? {
        log.info("Received notification ${message.payload}")
        val event = kotlin.runCatching { message.payload.mapTo(NotificationEvent::class.java) }
            .onFailure { log.error("Failed to parse notification event ${it.message}", it) }
            .getOrNull()
        return when (event) {
            is NotificationCreated -> handleNotificationCreated(message.withPayload(event))
            else -> {
                log.debug("Unhandled notification event")
                message.ack()
            }
        }
    }

    suspend fun handleNotificationCreated(event: Message<NotificationCreated>): CompletionStage<Void>? {
        return notificationService.sendNotification(event.payload.notificationId)
            .onSuccess {
                log.info("Successfully sent notification ${event.payload.notificationId}")
            }
            .onFailure {
                log.error("Failed to send notification ${event.payload.notificationId}", it)
            }
            .fold({ event.ack() }, { event.nack(it) })
    }
}