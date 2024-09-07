package it.beaesthetic.appointment.agenda.port.queue

import io.vertx.core.json.JsonObject
import it.beaesthetic.appointment.agenda.application.reminder.SendAgendaScheduleReminderHandler
import it.beaesthetic.appointment.agenda.application.reminder.SendReminder
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderTimesUp
import jakarta.enterprise.context.ApplicationScoped
import java.util.concurrent.CompletionStage
import org.eclipse.microprofile.reactive.messaging.Incoming
import org.eclipse.microprofile.reactive.messaging.Message
import org.jboss.logging.Logger

@ApplicationScoped
class SchedulerQueueConsumer(
    private val sendAgendaScheduleReminderHandler: SendAgendaScheduleReminderHandler
) {

    private val log = Logger.getLogger(SchedulerQueueConsumer::class.java)

    @Incoming("schedules")
    suspend fun handle(message: Message<JsonObject>): CompletionStage<Void>? {
        log.info("Received scheduler event ${message.payload}")
        val event =
            kotlin
                .runCatching { message.payload.mapTo(ReminderTimesUp::class.java) }
                .onFailure {
                    log.error("Failed to parse reminder times up event ${it.message}", it)
                }
                .getOrNull()
        return when (event) {
            is ReminderTimesUp ->
                sendAgendaScheduleReminderHandler
                    .handle(SendReminder(event.eventId))
                    .onSuccess { log.info("Successfully sent reminder ${it.id}") }
                    .onFailure { log.error("Failed to send reminder ${event.eventId}", it) }
                    .fold({ message.ack() }, { message.nack(it) })
            else -> {
                log.debug("Unhandled reminder event")
                message.ack()
            }
        }
    }
}
