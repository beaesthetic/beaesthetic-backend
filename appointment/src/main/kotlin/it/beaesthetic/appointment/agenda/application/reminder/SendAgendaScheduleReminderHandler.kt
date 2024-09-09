package it.beaesthetic.appointment.agenda.application.reminder

import arrow.core.flatMap
import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.event.CustomerRegistry
import it.beaesthetic.appointment.agenda.domain.notification.NotificationService
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import it.beaesthetic.appointment.agenda.domain.reminder.template.ReminderTemplateEngine
import jakarta.enterprise.context.ApplicationScoped
import org.jboss.logging.Logger

data class SendReminder(val eventId: AgendaEventId)

@ApplicationScoped
class SendAgendaScheduleReminderHandler(
    private val agendaRepository: AgendaRepository,
    private val customerRegistry: CustomerRegistry,
    private val notificationService: NotificationService,
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
            return kotlin
                .runCatching {
                    notificationService.trackAndSendReminderNotification(
                        event,
                        templateEngine,
                        customer.phoneNumber
                    )
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
