package it.beaesthetic.notification.driver.rest

import io.quarkus.runtime.annotations.RegisterForReflection
import it.beaesthetic.notification.application.NotificationService
import it.beaesthetic.notification.domain.Email
import it.beaesthetic.notification.domain.NotificationRepository
import it.beaesthetic.notification.domain.Sms
import it.beaesthetic.notification.domain.WhatsApp
import it.beaesthetic.notification.generated.api.NotificationsApi
import it.beaesthetic.notification.generated.api.model.*
import jakarta.ws.rs.BadRequestException
import jakarta.ws.rs.NotFoundException
import java.util.*

@RegisterForReflection(targets = [CreateNotification200ResponseDto::class, NotificationDto::class])
class NotificationResource(
    private val notificationService: NotificationService,
    private val notificationRepository: NotificationRepository,
) : NotificationsApi {

    override suspend fun createNotification(
        sendNotificationRequestDto: SendNotificationRequestDto
    ): CreateNotification200ResponseDto {
        val channel =
            when (sendNotificationRequestDto.channel) {
                is SmsChannelDto -> Sms(sendNotificationRequestDto.channel.phone)
                is WhatsappChannelDto -> WhatsApp(sendNotificationRequestDto.channel.phone)
                is EmailChannelDto -> Email(sendNotificationRequestDto.channel.email)
                else -> throw BadRequestException("Invalid channel type")
            }
        return notificationService
            .createNotification(
                sendNotificationRequestDto.title,
                sendNotificationRequestDto.content,
                channel
            )
            .map { CreateNotification200ResponseDto(UUID.fromString(it.id)) }
            .getOrThrow()
    }

    override suspend fun getNotification(notificationId: UUID): NotificationDto {
        return notificationRepository.findById(notificationId.toString())?.let {
            NotificationDto(
                notificationId = UUID.fromString(it.id),
                title = it.title,
                content = it.content,
                isSent = it.isSent,
                isSentConfirmed = it.isSentConfirmed,
                channel =
                    when (it.channel) {
                        is Email -> EmailChannelDto(it.channel.email)
                        is Sms -> SmsChannelDto(it.channel.phone)
                        is WhatsApp -> WhatsappChannelDto(it.channel.phone)
                    }
            )
        }
            ?: throw NotFoundException("Notification with id $notificationId not found")
    }
}
