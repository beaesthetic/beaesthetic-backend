package it.beaesthetic.appointment.agenda.infra

import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.domain.event.Customer
import it.beaesthetic.appointment.agenda.domain.event.CustomerRegistry
import it.beaesthetic.generated.customer.client.api.CustomersAdminApi

class RemoteCustomerRegistry(private val customersApi: CustomersAdminApi) : CustomerRegistry {
    override suspend fun findByCustomerId(customerId: String): Customer? {
        return customersApi
            .getCustomerById(customerId)
            .map {
                Customer(
                    customerId = it.id,
                    displayName = listOf(it.name, it.surname).joinToString(" "),
                    phoneNumber = it.phone,
                )
            }
            .onFailure()
            .recoverWithNull()
            .awaitSuspending()
    }
}
