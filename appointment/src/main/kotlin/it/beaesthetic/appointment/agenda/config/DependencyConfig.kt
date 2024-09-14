package it.beaesthetic.appointment.agenda.config

import com.fasterxml.jackson.databind.ObjectMapper
import io.quarkus.redis.datasource.ReactiveRedisDataSource
import it.beaesthetic.appointment.agenda.domain.Clock
import it.beaesthetic.appointment.agenda.domain.event.CustomerRegistry
import it.beaesthetic.appointment.agenda.domain.notification.NotificationService
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderOptions
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderScheduler
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderService
import it.beaesthetic.appointment.agenda.domain.reminder.template.ReminderTemplateEngine
import it.beaesthetic.appointment.agenda.infra.NotificationServiceImpl
import it.beaesthetic.appointment.agenda.infra.RemoteCustomerRegistry
import it.beaesthetic.appointment.agenda.infra.RemoteScheduler
import it.beaesthetic.appointment.agenda.infra.mongo.CancelReasonCodecProvider
import it.beaesthetic.generated.customer.client.api.CustomersApi
import it.beaesthetic.generated.notification.client.api.NotificationsApi
import it.beaesthetic.generated.scheduler.client.api.SchedulesApi
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces
import jakarta.inject.Singleton
import java.time.Duration
import org.eclipse.microprofile.config.inject.ConfigProperty
import org.eclipse.microprofile.rest.client.inject.RestClient

@Dependent
class DependencyConfig {

    @Produces
    fun notificationProvider(
        @RestClient customersApi: CustomersApi,
    ): CustomerRegistry {
        return RemoteCustomerRegistry(customersApi)
    }

    @Produces
    @Singleton
    fun cancelReasonCodecProvider(): CancelReasonCodecProvider {
        return CancelReasonCodecProvider()
    }

    @Produces
    @Singleton
    fun reminderScheduler(
        @RestClient schedulerApi: SchedulesApi,
        @ConfigProperty(name = "queues.scheduler-queue") schedulerRoute: String,
        objectMapper: ObjectMapper,
    ): ReminderScheduler {
        return RemoteScheduler(schedulerApi, schedulerRoute, objectMapper)
    }

    @Produces
    @Singleton
    fun notificationService(
        @RestClient notificationsApi: NotificationsApi,
        redis: ReactiveRedisDataSource,
    ): NotificationService {
        val notificationEventMap = redis.value(String::class.java)
        val notificationEventMapExpire = Duration.ofDays(1)
        return NotificationServiceImpl(
            notificationsApi,
            notificationEventMap,
            notificationEventMapExpire
        )
    }

    @Produces
    @Singleton
    fun reminderService(
        reminderConfiguration: ReminderConfiguration,
        notificationService: NotificationService,
        reminderScheduler: ReminderScheduler,
        reminderTemplateEngine: ReminderTemplateEngine
    ): ReminderService {
        return ReminderService(
            reminderScheduler = reminderScheduler,
            notificationService = notificationService,
            templateEngine = reminderTemplateEngine,
            reminderOptions =
                ReminderOptions(
                    reminderConfiguration.triggerBefore(),
                    reminderConfiguration.noSendThreshold(),
                    reminderConfiguration.immediateSendThreshold()
                ),
            clock = Clock.default()
        )
    }
}
