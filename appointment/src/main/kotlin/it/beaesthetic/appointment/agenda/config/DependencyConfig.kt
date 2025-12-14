package it.beaesthetic.appointment.agenda.config

import com.fasterxml.jackson.databind.ObjectMapper
import io.micrometer.core.instrument.MeterRegistry
import io.micrometer.core.instrument.Metrics
import io.opentelemetry.api.OpenTelemetry
import io.opentelemetry.instrumentation.micrometer.v1_5.OpenTelemetryMeterRegistry
import io.quarkus.redis.datasource.ReactiveRedisDataSource
import it.beaesthetic.appointment.agenda.application.reminder.ReminderTracker
import it.beaesthetic.appointment.agenda.domain.Clock
import it.beaesthetic.appointment.agenda.domain.event.CustomerRegistry
import it.beaesthetic.appointment.agenda.domain.notification.NotificationService
import it.beaesthetic.appointment.agenda.domain.notification.PendingNotification
import it.beaesthetic.appointment.agenda.domain.notification.template.NotificationTemplateEngine
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderOptions
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderScheduler
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderService
import it.beaesthetic.appointment.agenda.infra.NotificationServiceImpl
import it.beaesthetic.appointment.agenda.infra.RemoteCustomerRegistry
import it.beaesthetic.appointment.agenda.infra.RemoteScheduler
import it.beaesthetic.appointment.agenda.infra.mongo.CancelReasonCodecProvider
import it.beaesthetic.generated.customer.client.api.CustomersAdminApi
import it.beaesthetic.generated.notification.client.api.NotificationsApi
import it.beaesthetic.generated.scheduler.client.api.SchedulesApi
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces
import jakarta.inject.Singleton
import java.time.Duration
import java.util.concurrent.TimeUnit
import org.eclipse.microprofile.config.inject.ConfigProperty
import org.eclipse.microprofile.rest.client.inject.RestClient

@Dependent
class DependencyConfig {

    @Produces
    @Singleton
    fun openTelemetryMeterRegistry(openTelemetry: OpenTelemetry): MeterRegistry {
        return OpenTelemetryMeterRegistry.builder(openTelemetry)
            .setPrometheusMode(false)
            .setMicrometerHistogramGaugesEnabled(true)
            .setBaseTimeUnit(TimeUnit.MILLISECONDS)
            .setClock(io.micrometer.core.instrument.Clock.SYSTEM)
            .build()
            .apply { Metrics.addRegistry(this) }
    }

    @Produces
    fun notificationProvider(@RestClient customersApi: CustomersAdminApi): CustomerRegistry {
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
        notificationTemplateEngine: NotificationTemplateEngine,
        redis: ReactiveRedisDataSource,
    ): NotificationService {
        val notificationEventMap = redis.value(PendingNotification::class.java)
        val notificationEventMapExpire = Duration.ofDays(1)
        return NotificationServiceImpl(
            notificationsApi,
            notificationTemplateEngine,
            notificationEventMap,
            notificationEventMapExpire,
        )
    }

    @Produces
    @Singleton
    fun reminderService(
        reminderConfiguration: ReminderConfiguration,
        notificationService: NotificationService,
        reminderScheduler: ReminderScheduler,
        reminderTracker: ReminderTracker,
    ): ReminderService {
        return ReminderService(
            reminderScheduler = reminderScheduler,
            notificationService = notificationService,
            reminderOptions =
                ReminderOptions(
                    reminderConfiguration.triggerBefore(),
                    reminderConfiguration.noSendThreshold(),
                    reminderConfiguration.immediateSendThreshold(),
                ),
            clock = Clock.default(),
            reminderTracker = reminderTracker,
        )
    }
}
