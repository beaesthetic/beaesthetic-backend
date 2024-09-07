package it.beaesthetic.appointment.agenda

import com.fasterxml.jackson.databind.ObjectMapper
import it.beaesthetic.appointment.agenda.domain.event.CustomerRegistry
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderScheduler
import it.beaesthetic.appointment.agenda.infra.RemoteCustomerRegistry
import it.beaesthetic.appointment.agenda.infra.RemoteScheduler
import it.beaesthetic.appointment.agenda.infra.mongo.CancelReasonCodecProvider
import it.beaesthetic.generated.customer.client.api.CustomersApi
import it.beaesthetic.generated.scheduler.client.api.SchedulesApi
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces
import jakarta.inject.Singleton
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
        objectMapper: ObjectMapper,
    ): ReminderScheduler {
        return RemoteScheduler(schedulerApi, objectMapper)
    }
}
