package it.beaesthetic.appointment.agenda.application.notification

import it.beaesthetic.appointment.agenda.domain.event.CustomerRegistry
import it.beaesthetic.appointment.agenda.domain.notification.Notification
import it.beaesthetic.appointment.agenda.domain.notification.NotificationService
import jakarta.enterprise.context.ApplicationScoped
import org.jboss.logging.Logger

data class SendNotificationCommand(val notification: Notification)

// TODO: create a notificaiton handle to handle reminder
// TODO: unify notification and reminder logics
@ApplicationScoped
class SendNotificationHandler(
    private val customerRegistry: CustomerRegistry,
    private val notificationService: NotificationService
) {
    private val log = Logger.getLogger(SendNotificationHandler::class.java)

    suspend fun handle(command: SendNotificationCommand) {
        val notification = command.notification
        val customer = customerRegistry.findByCustomerId(notification.event.attendee.id)
        if (customer?.phoneNumber == null) {
            log.info(
                "Attendee ${customer?.customerId} has no valid contacts, not sending notification"
            )
        } else {
            notificationService
                .trackAndSendNotification(notification, customer.phoneNumber)
                .onSuccess { log.info("Sent notification for event ${notification.event.id}") }
                .onFailure {
                    log.error("Failed to send notification for event ${notification.event.id}", it)
                }
        }
    }
}
