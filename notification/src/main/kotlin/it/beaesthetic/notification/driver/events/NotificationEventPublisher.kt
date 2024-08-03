package it.beaesthetic.notification.driver.events

import io.quarkus.vertx.ConsumeEvent
import it.beaesthetic.notification.EventConfiguration
import it.beaesthetic.notification.domain.NotificationEvent
import it.beaesthetic.notification.domain.NotificationSentConfirmed
import jakarta.enterprise.context.ApplicationScoped
import org.eclipse.microprofile.reactive.messaging.Channel
import org.eclipse.microprofile.reactive.messaging.Emitter
import org.jboss.logging.Logger

@ApplicationScoped
class NotificationEventPublisher(
    @Channel("notifications-out")
    private val queueNotificationEventEmitter: Emitter<NotificationEvent>,
    @Channel("notifications-confirmed-out")
    private val queueNotificationConfirmedEventEmitter: Emitter<NotificationEvent>
) {

    private val log = Logger.getLogger(NotificationEventPublisher::class.java)

    @ConsumeEvent(EventConfiguration.INTERNAL_DOMAIN_EVENTS_TOPIC)
    fun handle(event: NotificationEvent) {
        when (event) {
            is NotificationSentConfirmed -> queueNotificationConfirmedEventEmitter.send(event)
            else -> queueNotificationEventEmitter.send(event)
        }.whenComplete { _, e ->
            if (e != null) log.error(e.message, e)
            else log.info("Successfully published notification event $event")
        }
    }
}
