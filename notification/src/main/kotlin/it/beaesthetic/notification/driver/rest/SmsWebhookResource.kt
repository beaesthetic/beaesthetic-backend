package it.beaesthetic.notification.driver.rest

import it.beaesthetic.notification.application.NotificationService
import it.beaesthetic.notification.infra.providers.SmsNotificationProvider
import it.beaesthetic.notification.sms.generated.api.SmsGatewayApi
import it.beaesthetic.notification.sms.generated.api.model.EventNotificationDtoDto
import it.beaesthetic.notification.sms.generated.api.model.EventNotificationTypeDto
import jakarta.ws.rs.BadRequestException
import org.jboss.logging.Logger

class SmsWebhookResource(private val notificationService: NotificationService) : SmsGatewayApi {

    private val log = Logger.getLogger(SmsWebhookResource::class.java)

    override suspend fun smsGatewayNotify(eventNotificationDtoDto: EventNotificationDtoDto) {
        when (eventNotificationDtoDto.eventType) {
            EventNotificationTypeDto.MESSAGE_PERIOD_DELIVER_PERIOD_SUCCEEDED -> {
                log.info(
                    "Received SMS delivery success event for ${eventNotificationDtoDto.data?.id}"
                )
                val notificationId =
                    eventNotificationDtoDto.metadata?.get(
                        SmsNotificationProvider.NOTIFICATION_ID_METADATA
                    )
                if (notificationId == null) {
                    log.error(
                        "Missing mandatory metadata [${SmsNotificationProvider.NOTIFICATION_ID_METADATA}]"
                    )
                    throw BadRequestException("Missing mandatory metadata")
                }

                notificationService.confirmNotificationSent(notificationId).getOrThrow()
            }
            EventNotificationTypeDto.MESSAGE_PERIOD_DELIVER_PERIOD_FAILED -> {
                log.info(
                    "Received SMS delivery failure event for ${eventNotificationDtoDto.data?.id}. Doing nothing"
                )
            }
            else -> {
                log.warn("Unexpected event type [${eventNotificationDtoDto.eventType}]")
            }
        }
    }
}
