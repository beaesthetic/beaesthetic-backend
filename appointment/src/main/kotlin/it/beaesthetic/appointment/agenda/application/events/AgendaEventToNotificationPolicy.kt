package it.beaesthetic.appointment.agenda.application.events

import io.quarkus.vertx.ConsumeEvent
import it.beaesthetic.appointment.agenda.application.notification.SendNotificationCommand
import it.beaesthetic.appointment.agenda.application.notification.SendNotificationHandler
import it.beaesthetic.appointment.agenda.config.NotificationConfiguration
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventRescheduled
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventScheduled
import it.beaesthetic.appointment.agenda.domain.notification.Notification
import it.beaesthetic.appointment.agenda.domain.notification.NotificationType
import it.beaesthetic.appointment.service.common.uniWithScope
import jakarta.enterprise.context.ApplicationScoped
import org.jboss.logging.Logger

@ApplicationScoped
class AgendaEventToNotificationPolicy(
    private val sendNotification: SendNotificationHandler
) {
    private val log = Logger.getLogger(AgendaEventToNotificationPolicy::class.java)

    @ConsumeEvent("AgendaEventScheduled")
    fun handle(event: AgendaEventScheduled) = uniWithScope {
        log.info("Notification policy handle for new agenda scheduled event, sending notification")
        sendNotification.handle(
            SendNotificationCommand(
                Notification(NotificationType.Confirmation(), event.agendaEvent)
            )
        )
    }

    @ConsumeEvent("AgendaEventRescheduled")
    fun handle(event: AgendaEventRescheduled) = uniWithScope {
        log.info("Notification policy handle for agenda re-scheduled event, sending notification")
        sendNotification.handle(
            SendNotificationCommand(
                Notification(NotificationType.Confirmation(true), event.agendaEvent)
            )
        )
    }
}
