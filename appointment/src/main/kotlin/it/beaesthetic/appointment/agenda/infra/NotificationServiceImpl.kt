package it.beaesthetic.appointment.agenda.infra

import io.quarkus.redis.datasource.value.ReactiveValueCommands
import io.quarkus.redis.datasource.value.SetArgs
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.notification.NotificationId
import it.beaesthetic.appointment.agenda.domain.notification.NotificationService
import it.beaesthetic.appointment.agenda.domain.reminder.template.ReminderTemplateEngine
import it.beaesthetic.generated.notification.client.api.NotificationsApi
import it.beaesthetic.generated.notification.client.model.NotificationChannel
import it.beaesthetic.generated.notification.client.model.SendNotificationRequest
import java.time.Duration

class NotificationServiceImpl(
    private val notificationsApi: NotificationsApi,
    private val eventNotificationMap: ReactiveValueCommands<String, String>,
    private val defaultEventNotificationMapTTL: Duration
) : NotificationService {

    override suspend fun trackAndSendReminderNotification(
        event: AgendaEvent,
        templateEngine: ReminderTemplateEngine,
        phoneNumber: String
    ): Result<NotificationId> {
        val body = templateEngine.process(event)
        val request =
            SendNotificationRequest().apply {
                title = ""
                content = body
                channel =
                    NotificationChannel().type(NotificationChannel.TypeEnum.SMS).phone(phoneNumber)
            }
        return runCatching {
            notificationsApi
                .createNotification(request)
                .flatMap { response ->
                    eventNotificationMap
                        .set(
                            response.notificationId.toString(),
                            event.id.value,
                            SetArgs().px(defaultEventNotificationMapTTL)
                        )
                        .map { NotificationId(response.notificationId.toString()) }
                }
                .awaitSuspending()
        }
    }

    override suspend fun findEventByNotification(
        notificationId: NotificationId
    ): Result<AgendaEventId> = runCatching {
        eventNotificationMap.get(notificationId.value).map { AgendaEventId(it) }.awaitSuspending()
    }

    override suspend fun removeTrackNotification(notificationId: NotificationId) {
        eventNotificationMap.getdel(notificationId.value).awaitSuspending()
    }
}
