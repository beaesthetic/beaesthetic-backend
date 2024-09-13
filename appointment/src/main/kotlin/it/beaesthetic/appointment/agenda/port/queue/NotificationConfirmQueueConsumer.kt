package it.beaesthetic.appointment.agenda.port.queue

import arrow.core.flatMap
import io.vertx.core.json.JsonObject
import it.beaesthetic.appointment.agenda.application.reminder.ConfirmReminderSent
import it.beaesthetic.appointment.agenda.application.reminder.ConfirmReminderSentHandler
import it.beaesthetic.appointment.agenda.domain.notification.NotificationId
import it.beaesthetic.appointment.agenda.domain.notification.NotificationService
import jakarta.enterprise.context.ApplicationScoped
import java.util.concurrent.CompletionStage
import org.eclipse.microprofile.reactive.messaging.Incoming
import org.eclipse.microprofile.reactive.messaging.Message
import org.jboss.logging.Logger

@ApplicationScoped
class NotificationConfirmQueueConsumer(
    private val notificationService: NotificationService,
    private val confirmReminderSentHandler: ConfirmReminderSentHandler
) {

    private val log = Logger.getLogger(NotificationConfirmQueueConsumer::class.java)

    @Incoming("notifications-confirmed")
    suspend fun handle(message: Message<JsonObject>): CompletionStage<Void>? {
        log.info("Received notification confirm event ${message.payload}")
        val notificationId = message.payload.getString("notificationId")?.let { NotificationId(it) }
        if (notificationId != null) {
            notificationService
                .findEventByNotification(notificationId)
                .onSuccess {
                    log.info("Found event associated to notification id ${notificationId}")
                }
                .flatMap { agendaEventId ->
                    confirmReminderSentHandler
                        .handle(ConfirmReminderSent(agendaEventId))
                        .onSuccess { log.info("Successfully confirm reminder for ${it.id}") }
                        .onFailure {
                            log.error("Failed to confirm reminder for $agendaEventId", it)
                        }
                }
                .map { notificationService.removeTrackNotification(notificationId) }
                .onFailure {
                    log.error("Failed to process confirm notification $notificationId", it)
                }
        } else {
            log.warn("Notification message doesn't contains notificationId")
        }
        return message.ack()
    }
}
