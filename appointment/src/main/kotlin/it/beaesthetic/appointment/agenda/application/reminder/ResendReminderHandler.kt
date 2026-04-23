package it.beaesthetic.appointment.agenda.application.reminder

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.event.CustomerRegistry
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderService
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import jakarta.enterprise.context.ApplicationScoped
import org.jboss.logging.Logger

data class ResendReminder(val eventId: AgendaEventId)

@ApplicationScoped
class ResendReminderHandler(
    private val agendaRepository: AgendaRepository,
    private val customerRegistry: CustomerRegistry,
    private val reminderService: ReminderService,
) {

    private val log = Logger.getLogger(ResendReminderHandler::class.java)

    suspend fun handle(command: ResendReminder): Result<AgendaEvent> {
        log.info("Resending reminder for event ${command.eventId}")
        val (event, version) =
            agendaRepository.findEvent(command.eventId)
                ?: throw IllegalArgumentException("No event with id ${command.eventId} found")

        if (event.cancelReason != null) {
            log.info("Cannot resend reminder for cancelled event ${command.eventId}")
            return Result.failure(IllegalArgumentException("Cannot resend reminder for cancelled event"))
        }

        val resetEvent = event.updateReminder(event.reminder.copy(status = ReminderStatus.SCHEDULED))
        val customer = customerRegistry.findByCustomerId(event.attendee.id)

        if (customer?.phoneNumber == null) {
            log.info("Customer ${customer?.customerId} has no phone number, cannot resend reminder")
            return Result.failure(IllegalArgumentException("Customer has no phone number"))
        }

        return reminderService
            .sendReminder(resetEvent, customer.phoneNumber)
            .let { agendaRepository.saveEvent(it, version) }
            .onSuccess {
                when (it.reminder.status) {
                    ReminderStatus.SENT_REQUESTED ->
                        log.info("Successfully resent reminder to notification service")
                    else ->
                        log.error(
                            "Failed to resend reminder cause status is ${it.reminder.status}"
                        )
                }
            }
            .onFailure { log.error("Failed to save resent reminder status", it) }
    }
}
