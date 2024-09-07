package it.beaesthetic.appointment.agenda.port.queue

import io.vertx.core.json.JsonObject
import it.beaesthetic.appointment.agenda.application.reminder.ConfirmReminderSentHandler
import jakarta.enterprise.context.ApplicationScoped
import java.util.concurrent.CompletionStage
import org.eclipse.microprofile.reactive.messaging.Incoming
import org.eclipse.microprofile.reactive.messaging.Message
import org.jboss.logging.Logger

@ApplicationScoped
class NotificationConfirmQueueConsumer(
    private val confirmReminderSentHandler: ConfirmReminderSentHandler
) {

    private val log = Logger.getLogger(NotificationConfirmQueueConsumer::class.java)

    @Incoming("notifications-confirmed")
    suspend fun handle(message: Message<JsonObject>): CompletionStage<Void>? {
        log.info("Received notification confirm event ${message.payload}")
        // confirmReminderSentHandler.handle(ConfirmReminderSent())
        //        val event =
        //            kotlin
        //                .runCatching { message.payload.mapTo(ReminderTimesUp::class.java) }
        //                .onFailure { log.error("Failed to parse reminder times up event
        // ${it.message}", it) }
        //                .getOrNull()
        //        return when (event) {
        //            is ReminderTimesUp ->
        // sendAgendaScheduleReminderHandler.handle(SendReminder(event.eventId))
        //                .onSuccess { log.info("Successfully sent reminder ${it.id}") }
        //                .onFailure { log.error("Failed to send reminder ${event.eventId}", it) }
        //                .fold({ message.ack() }, { message.nack(it) })
        //            else -> {
        //                log.debug("Unhandled reminder event")
        //                message.ack()
        //            }
        //        }
        return message.ack()
    }
}
