package it.beaesthetic.appointment.agenda.application.reminder

import arrow.core.flatMap
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.event.CustomerRegistry
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import it.beaesthetic.appointment.agenda.domain.reminder.template.ReminderTemplateEngine
import it.beaesthetic.generated.notification.client.api.NotificationsApi
import it.beaesthetic.generated.notification.client.model.NotificationChannel
import it.beaesthetic.generated.notification.client.model.SendNotificationRequest
import jakarta.enterprise.context.ApplicationScoped
import org.jboss.logging.Logger

data class SendReminder(val eventId: String)

@ApplicationScoped
class SendAgendaScheduleReminderHandler(
    private val agendaRepository: AgendaRepository,
    private val customerRegistry: CustomerRegistry,
    private val notificationApi: NotificationsApi,
    private val templateEngine: ReminderTemplateEngine
) {

    private val log = Logger.getLogger(SendAgendaScheduleReminderHandler::class.java)

    suspend fun handle(command: SendReminder): Result<AgendaEvent> {
        val (event, version) =
            agendaRepository.findEvent(command.eventId)
                ?: throw IllegalArgumentException("No event with id ${command.eventId} found")

        val customer = customerRegistry.findByCustomerId(event.attendee.id)
        if (customer?.phoneNumber == null) {
            log.info("Customer ${customer?.customerId} has no valid contacts, not sending reminder")
            return agendaRepository.saveEvent(
                event.updateReminderStatus(ReminderStatus.SENT),
                version
            )
        } else {
            val message = templateEngine.process(event)
            val notificationRequest =
                SendNotificationRequest().apply {
                    title = "Reminder of ${command.eventId} for ${customer.customerId}"
                    content = message
                    channel =
                        NotificationChannel()
                            .type(NotificationChannel.TypeEnum.SMS)
                            .phone(customer.phoneNumber)
                }

            return kotlin
                .runCatching {
                    notificationApi.createNotification(notificationRequest).awaitSuspending()
                }
                .onSuccess { log.info("Successfully sent reminder to notification service") }
                .onFailure { log.error("Failed to sent reminder to notification service", it) }
                .flatMap {
                    agendaRepository.saveEvent(
                        event.updateReminderStatus(ReminderStatus.SENT_REQUESTED),
                        version
                    )
                }
        }
    }
}
