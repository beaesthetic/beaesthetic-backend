package it.beaesthetic.appointment.agenda.infra

import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.domain.Customer
import it.beaesthetic.appointment.agenda.domain.CustomerRegistry
import it.beaesthetic.generated.customer.client.api.CustomersApi
import jakarta.enterprise.context.ApplicationScoped

class RemoteCustomerRegistry(
        private val customersApi: CustomersApi
) : CustomerRegistry {
    override suspend fun findByCustomerId(customerId: String): Customer? {
        return customersApi.getCustomerById(customerId)
                .map { Customer(
                        customerId = it.id,
                        displayName = listOf(it.name, it.surname).joinToString(" ")
                ) }
                .onFailure()
                .recoverWithNull()
                .awaitSuspending()
    }
}