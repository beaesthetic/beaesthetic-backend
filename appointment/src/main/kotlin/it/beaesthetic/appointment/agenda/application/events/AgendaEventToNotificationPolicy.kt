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
    notificationConfiguration: NotificationConfiguration,
    private val sendNotification: SendNotificationHandler
) {
    private val log = Logger.getLogger(AgendaEventToNotificationPolicy::class.java)

    private val customerWhitelist = notificationConfiguration.whitelist().toSet()

    @ConsumeEvent("AgendaEventScheduled")
    fun handle(event: AgendaEventScheduled) = uniWithScope {
        log.info("Notification policy handle for new agenda scheduled event")
        if (!customerWhitelist.contains(event.agendaEvent.attendee.id)) {
            log.info(
                "Attendee ${event.agendaEvent.attendee.id} not in whitelist, skipping notification"
            )
            return@uniWithScope
        }
        log.info("Attendee in whitelist, sending notification")
        sendNotification.handle(
            SendNotificationCommand(
                Notification(NotificationType.Confirmation(), event.agendaEvent)
            )
        )
    }

    @ConsumeEvent("AgendaEventRescheduled")
    fun handle(event: AgendaEventRescheduled) = uniWithScope {
        log.info("Notification policy handle for agenda re-scheduled event")
        if (!customerWhitelist.contains(event.agendaEvent.attendee.id)) {
            log.info(
                "Attendee ${event.agendaEvent.attendee.id} not in whitelist, skipping notification"
            )
            return@uniWithScope
        }
        log.info("Attendee in whitelist, sending notification")
        sendNotification.handle(
            SendNotificationCommand(
                Notification(NotificationType.Confirmation(true), event.agendaEvent)
            )
        )
    }
}
