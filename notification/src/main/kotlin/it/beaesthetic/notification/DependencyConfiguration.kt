package it.beaesthetic.notification

import io.quarkus.runtime.annotations.RegisterForReflection
import it.beaesthetic.generated.smsgateway.api.SmsApi
import it.beaesthetic.notification.application.NotificationService
import it.beaesthetic.notification.configmapping.SmsGatewayConfig
import it.beaesthetic.notification.domain.*
import it.beaesthetic.notification.infra.providers.SmsNotificationProvider
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces
import jakarta.inject.Singleton
import org.bson.codecs.configuration.CodecProvider
import org.bson.codecs.pojo.ClassModel
import org.bson.codecs.pojo.PojoCodecProvider
import org.eclipse.microprofile.rest.client.inject.RestClient
import java.util.UUID

@Dependent
class DependencyConfiguration {

    @Produces
    fun notificationService(
        notificationRepository: NotificationRepository,
        notificationProvider: NotificationProvider
    ): NotificationService {
        return NotificationService(notificationRepository, notificationProvider)
    }

    @Produces
    fun notificationProvider(
        @RestClient smsApi: SmsApi,
        smsGatewayConfig: SmsGatewayConfig
    ): NotificationProvider {
        return CompoundNotificationProvider(
            listOf(
                object : NotificationProvider {
                    override suspend fun isSupported(notification: Notification): Boolean = true

                    override suspend fun send(notification: Notification): Result<ChannelMetadata> =
                        Result.success(ChannelMetadata(providerResourceId = UUID.randomUUID().toString()))

                },
                //SmsNotificationProvider(smsApi, smsGatewayConfig.senderNumber())
            )
        )
    }

    @Singleton
    fun registerPojoCodes(): CodecProvider {
        return PojoCodecProvider.builder()
            .register(
                ClassModel.builder(Channel::class.java)
                    .enableDiscriminator(true)
                    .discriminatorKey("type")
                    .build(),
                ClassModel.builder(Sms::class.java)
                    .enableDiscriminator(true)
                    .discriminatorKey("type")
                    .discriminator("sms")
                    .build()
            )
            .build()
    }
}