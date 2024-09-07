package it.beaesthetic.appointment.agenda.application.events

import io.quarkus.vertx.ConsumeEvent
import it.beaesthetic.appointment.agenda.application.reminder.DeleteReminder
import it.beaesthetic.appointment.agenda.application.reminder.DeleteReminderHandler
import it.beaesthetic.appointment.agenda.application.reminder.ScheduleReminder
import it.beaesthetic.appointment.agenda.application.reminder.ScheduleReminderHandler
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventDeleted
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventRescheduled
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventScheduled
import it.beaesthetic.appointment.service.common.uniWithScope
import jakarta.enterprise.context.ApplicationScoped
import org.jboss.logging.Logger

@ApplicationScoped
class AgendaEventToReminderPolicy(
    private val scheduleReminderHandler: ScheduleReminderHandler,
    private val deleteReminderHandler: DeleteReminderHandler,
) {
    private val log = Logger.getLogger(AgendaEventToReminderPolicy::class.java)

    @ConsumeEvent("AgendaEventScheduled")
    fun handle(event: AgendaEventScheduled) = uniWithScope {
        log.info("Reminder policy handle for new agenda scheduled event")
        val command = ScheduleReminder(event.agendaEvent.id)
        scheduleReminderHandler
            .handle(command)
            .onSuccess { log.info("Successfully handled schedule reminder command") }
            .onFailure { log.error("Failed to schedule agenda scheduled event", it) }
    }

    @ConsumeEvent("AgendaEventRescheduled")
    fun handle(event: AgendaEventRescheduled) = uniWithScope {
        log.info("Reminder policy handle for agenda re-scheduled event")
        val command = ScheduleReminder(event.agendaEvent.id)
        scheduleReminderHandler
            .handle(command)
            .onSuccess { log.info("Successfully handled schedule reminder command") }
            .onFailure { log.error("Failed to re-schedule agenda scheduled event", it) }
    }

    @ConsumeEvent("AgendaEventDeleted")
    fun handle(event: AgendaEventDeleted) = uniWithScope {
        val command = DeleteReminder(event.agendaEvent.id)
        deleteReminderHandler.handle(command)
    }
}
