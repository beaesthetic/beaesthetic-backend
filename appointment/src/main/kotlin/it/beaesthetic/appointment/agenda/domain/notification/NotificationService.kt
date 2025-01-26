package it.beaesthetic.appointment.agenda.domain.notification

import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId

@JvmInline value class NotificationId(val value: String)

data class PendingNotification(
    val notificationId: NotificationId,
    val agendaEventId: AgendaEventId,
    val notificationType: NotificationType
)

interface NotificationService {
    suspend fun trackAndSendNotification(
        notification: Notification,
        phoneNumber: String
    ): Result<PendingNotification>

    suspend fun findPendingNotification(notificationId: NotificationId): Result<PendingNotification>

    suspend fun removeTrackNotification(notificationId: NotificationId)
}
