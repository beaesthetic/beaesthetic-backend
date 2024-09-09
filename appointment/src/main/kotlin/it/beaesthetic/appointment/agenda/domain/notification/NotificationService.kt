package it.beaesthetic.appointment.agenda.domain.notification

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.reminder.template.ReminderTemplateEngine

@JvmInline value class NotificationId(val value: String)

interface NotificationService {
    suspend fun trackAndSendReminderNotification(
        event: AgendaEvent,
        templateEngine: ReminderTemplateEngine,
        phoneNumber: String
    ): Result<NotificationId>

    suspend fun findEventByNotification(notificationId: NotificationId): Result<AgendaEventId>

    suspend fun removeTrackNotification(notificationId: NotificationId)
}
