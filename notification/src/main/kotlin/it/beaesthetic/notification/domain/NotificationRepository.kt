package it.beaesthetic.notification.domain

interface NotificationRepository {
    suspend fun findById(notificationId: String): Notification?
    suspend fun save(notification: Notification): Result<Notification>
}