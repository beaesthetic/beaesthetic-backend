package it.beaesthetic.notification

import it.beaesthetic.generated.smsgateway.api.SmsApi
import it.beaesthetic.notification.application.NotificationService
import it.beaesthetic.notification.configmapping.SmsGatewayConfig
import it.beaesthetic.notification.domain.*
import it.beaesthetic.notification.infra.providers.SmsNotificationProvider
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces
import jakarta.inject.Singleton
import jakarta.ws.rs.core.Response
import jakarta.ws.rs.ext.ContextResolver
import jakarta.ws.rs.ext.Provider
import org.bson.codecs.configuration.CodecProvider
import org.bson.codecs.pojo.ClassModel
import org.bson.codecs.pojo.PojoCodecProvider
import org.eclipse.microprofile.rest.client.inject.RestClient
import org.jboss.resteasy.reactive.client.handlers.RedirectHandler

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
            listOf(SmsNotificationProvider(smsApi, smsGatewayConfig.senderNumber()))
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

    /** Allows to follow redirection of sms gateway rest client */
    @Provider
    class AlwaysRedirectHandler : ContextResolver<RedirectHandler> {
        override fun getContext(aClass: Class<*>?): RedirectHandler {
            return RedirectHandler { response: Response ->
                if (
                    Response.Status.Family.familyOf(response.status) ===
                        Response.Status.Family.REDIRECTION
                ) {
                    return@RedirectHandler response.location
                }
                null
            }
        }
    }
}
