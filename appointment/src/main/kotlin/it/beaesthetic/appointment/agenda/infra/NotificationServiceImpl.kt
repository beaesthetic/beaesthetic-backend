package it.beaesthetic.appointment.agenda.infra

import arrow.core.flatMap
import com.fasterxml.jackson.annotation.JsonSubTypes
import com.fasterxml.jackson.annotation.JsonTypeInfo
import io.quarkus.redis.datasource.value.ReactiveValueCommands
import io.quarkus.redis.datasource.value.SetArgs
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.domain.notification.*
import it.beaesthetic.appointment.agenda.domain.notification.template.NotificationTemplateEngine
import it.beaesthetic.generated.notification.client.api.NotificationsApi
import it.beaesthetic.generated.notification.client.model.NotificationChannel
import it.beaesthetic.generated.notification.client.model.SendNotificationRequest
import java.time.Duration

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, property = "type")
@JsonSubTypes(
    value =
        [
            JsonSubTypes.Type(value = NotificationType.Reminder::class, name = "reminder"),
            JsonSubTypes.Type(value = NotificationType.Confirmation::class, name = "confirmation"),
        ]
)
interface NotificationTypeMixin

class NotificationServiceImpl(
    private val notificationsApi: NotificationsApi,
    private val templateEngine: NotificationTemplateEngine,
    private val eventNotificationMap: ReactiveValueCommands<String, PendingNotification>,
    private val defaultEventNotificationMapTTL: Duration,
) : NotificationService {

    companion object {
        private const val REDIS_PREFIX = "pending-notification"

        private fun formatRedisKey(notificationId: NotificationId): String {
            return "${REDIS_PREFIX}:${notificationId.value}"
        }
    }

    override suspend fun trackAndSendNotification(
        notification: Notification,
        phoneNumber: String,
    ): Result<PendingNotification> {
        return templateEngine
            .process(notification)
            .map { body ->
                SendNotificationRequest().apply {
                    title = ""
                    content = body
                    channel =
                        NotificationChannel()
                            .type(NotificationChannel.TypeEnum.SMS)
                            .phone(phoneNumber)
                }
            }
            .flatMap { request ->
                runCatching {
                    notificationsApi
                        .createNotification(request)
                        .map {
                            PendingNotification(
                                NotificationId(it.notificationId.toString()),
                                notification.event.id,
                                notification.type,
                            )
                        }
                        .flatMap { pendingNotification ->
                            eventNotificationMap
                                .set(
                                    formatRedisKey(pendingNotification.notificationId),
                                    pendingNotification,
                                    SetArgs().px(defaultEventNotificationMapTTL),
                                )
                                .map { pendingNotification }
                        }
                        .awaitSuspending()
                }
            }
    }

    override suspend fun findPendingNotification(
        notificationId: NotificationId
    ): Result<PendingNotification> = runCatching {
        eventNotificationMap.get(formatRedisKey(notificationId)).awaitSuspending()
    }

    override suspend fun removeTrackNotification(notificationId: NotificationId) {
        eventNotificationMap.getdel(formatRedisKey(notificationId)).awaitSuspending()
    }
}
