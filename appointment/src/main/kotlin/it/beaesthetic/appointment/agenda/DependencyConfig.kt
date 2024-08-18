package it.beaesthetic.appointment.agenda

import it.beaesthetic.appointment.agenda.domain.CustomerRegistry
import it.beaesthetic.appointment.agenda.infra.RemoteCustomerRegistry
import it.beaesthetic.generated.customer.client.api.CustomersApi
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces
import org.eclipse.microprofile.rest.client.inject.RestClient

@Dependent
class DependencyConfig {

    @Produces
    fun notificationProvider(
            @RestClient customersApi: CustomersApi,
    ): CustomerRegistry {
        return RemoteCustomerRegistry(customersApi)
    }
}