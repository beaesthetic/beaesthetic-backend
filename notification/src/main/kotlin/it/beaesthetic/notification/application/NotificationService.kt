package it.beaesthetic.notification.application

import arrow.core.flatMap
import it.beaesthetic.notification.domain.Channel
import it.beaesthetic.notification.domain.Notification
import it.beaesthetic.notification.domain.NotificationProvider
import it.beaesthetic.notification.domain.NotificationRepository
import java.util.UUID

class NotificationService(
    private val notificationRepository: NotificationRepository,
    private val notificationProvider: NotificationProvider
) {

    suspend fun createNotification(
        title: String,
        content: String,
        channel: Channel
    ): Result<Notification> {
        val notification =
            Notification.of(
                id = UUID.randomUUID().toString(),
                title = title,
                content = content,
                channel = channel
            )
        return notificationRepository.save(notification)
    }

    suspend fun sendNotification(notificationId: String): Result<Notification> {
        val notification =
            notificationRepository.findById(notificationId)
                ?: return Result.failure(
                    IllegalArgumentException("Notification $notificationId not found")
                )

        if (notificationProvider.isSupported(notification)) {
            return notificationProvider
                .send(notification)
                .map { notification.markSendWith(it) }
                .flatMap { notificationRepository.save(it) }
        } else {
            return Result.failure(
                NoSuchElementException("Provided for ${notification.channel} is not supported")
            )
        }
    }

    suspend fun confirmNotificationSent(notificationId: String): Result<Notification> {
        val notification =
            notificationRepository.findById(notificationId)
                ?: return Result.failure(
                    IllegalArgumentException("Notification $notificationId not found")
                )
        return notificationRepository.save(notification.confirmSend())
    }
}
