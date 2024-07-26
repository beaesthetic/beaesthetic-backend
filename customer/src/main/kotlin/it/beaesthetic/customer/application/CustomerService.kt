package it.beaesthetic.customer.application

import it.beaesthetic.customer.domain.*
import it.beaesthetic.customer.generated.api.model.CustomerCreateDto
import it.beaesthetic.customer.generated.api.model.CustomerUpdateDto

class CustomerService(private val customerRepository: CustomerRepository) {

    suspend fun createCustomer(createRequest: CustomerCreateDto): Customer {
        val contacts =
            Contacts(createRequest.email?.let { Email(it) }, createRequest.phone?.let { Phone(it) })
        val customer =
            Customer.create(
                name = createRequest.name,
                surname = createRequest.surname ?: "",
                contacts = contacts,
                note = createRequest.note ?: ""
            )
        return customerRepository.save(customer)
    }

    suspend fun updateCustomer(customerId: CustomerId, updateDto: CustomerUpdateDto): Customer? {
        val customer =
            customerRepository
                .findById(customerId)
                ?.let {
                    it.copy(
                        name = updateDto.name ?: it.name,
                        surname = updateDto.surname ?: it.surname,
                        note = updateDto.note ?: it.note,
                    )
                }
                ?.let {
                    it.changeContacts(
                        Contacts(
                            email = updateDto.email?.let { v -> Email(v) } ?: it.contacts.email,
                            phone = updateDto.phone?.let { v -> Phone(v) } ?: it.contacts.phone,
                        )
                    )
                }

        return customer?.let { customerRepository.save(customer) }
    }
}
