package it.beaesthetic.customer

import it.beaesthetic.customer.application.CustomerService
import it.beaesthetic.customer.domain.CustomerRepository
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces

@Dependent
class DependencyConfiguration {

    @Produces
    fun customerService(customerRepository: CustomerRepository): CustomerService {
        return CustomerService(customerRepository)
    }
}
