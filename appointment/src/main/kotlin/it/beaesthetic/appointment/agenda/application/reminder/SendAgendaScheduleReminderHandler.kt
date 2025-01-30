package it.beaesthetic.appointment.agenda.application.reminder

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.event.CustomerRegistry
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderService
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import jakarta.enterprise.context.ApplicationScoped
import kotlin.onSuccess
import org.jboss.logging.Logger

data class SendReminder(val eventId: AgendaEventId)

@ApplicationScoped
class SendAgendaScheduleReminderHandler(
    private val agendaRepository: AgendaRepository,
    private val customerRegistry: CustomerRegistry,
    private val reminderService: ReminderService,
) {

    private val log = Logger.getLogger(SendAgendaScheduleReminderHandler::class.java)

    suspend fun handle(command: SendReminder): Result<AgendaEvent> {
        val (event, version) =
            agendaRepository.findEvent(command.eventId)
                ?: throw IllegalArgumentException("No event with id ${command.eventId} found")

        val customer = customerRegistry.findByCustomerId(event.attendee.id)
        if (customer?.phoneNumber == null) {
            log.info("Customer ${customer?.customerId} has no valid contacts, not sending reminder")
            return Result.success(event)
        } else {
            return reminderService
                .sendReminder(event, customer.phoneNumber)
                .let { agendaRepository.saveEvent(it, version) }
                .onSuccess {
                    when (it.reminder.status) {
                        ReminderStatus.SENT_REQUESTED ->
                            log.info("Successfully sent reminder to notification service")
                        else ->
                            log.error(
                                "Failed to sent reminder cause status is ${event.reminder.status}"
                            )
                    }
                }
                .onFailure { log.error("Failed to save reminder status", it) }
        }
    }
}
