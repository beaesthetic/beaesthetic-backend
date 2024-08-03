package it.beaesthetic.notification.infra.providers

import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.generated.smsgateway.api.SmsApi
import it.beaesthetic.generated.smsgateway.model.SendSmsRequest
import it.beaesthetic.notification.domain.ChannelMetadata
import it.beaesthetic.notification.domain.Notification
import it.beaesthetic.notification.domain.NotificationProvider
import it.beaesthetic.notification.domain.Sms
import org.jboss.logging.Logger

class SmsNotificationProvider(private val smsApi: SmsApi, private val fromNumber: String) :
    NotificationProvider {

    private val log = Logger.getLogger(SmsNotificationProvider::class.java)

    companion object {
        const val NOTIFICATION_ID_METADATA = "notificationId"
    }

    override suspend fun isSupported(notification: Notification): Boolean =
        notification.channel is Sms

    override suspend fun send(notification: Notification): Result<ChannelMetadata> =
        runCatching {
                val channel = notification.channel as Sms
                val response =
                    smsApi
                        .sendSms(
                            notification.id,
                            SendSmsRequest()
                                .to(channel.phone)
                                .from(fromNumber)
                                .content(notification.content)
                                .putMetadataItem(NOTIFICATION_ID_METADATA, notification.id)
                        )
                        .awaitSuspending()

                ChannelMetadata(providerResourceId = response.id.toString())
            }
            .onSuccess { log.info("Successfully send sms notification ${it.providerResourceId}") }
            .onFailure { log.error("Error sending sms notification ${it.message}", it) }
}
