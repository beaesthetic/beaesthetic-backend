package it.beaesthetic.customer.rest

import it.beaesthetic.customer.application.CustomerService
import it.beaesthetic.customer.domain.Customer
import it.beaesthetic.customer.domain.CustomerId
import it.beaesthetic.customer.domain.CustomerRepository
import it.beaesthetic.customer.generated.api.CustomersApi
import it.beaesthetic.customer.generated.api.model.*
import it.beaesthetic.customer.rest.CustomerController.ResourceMapper.toResource
import jakarta.ws.rs.NotFoundException

class CustomerController(
    private val customerService: CustomerService,
    private val customerRepository: CustomerRepository
) : CustomersApi {

    override suspend fun createCustomer(customerCreateDto: CustomerCreateDto): CreateCustomer201ResponseDto {
        return CreateCustomer201ResponseDto(id = customerService.createCustomer(customerCreateDto).id.value)
    }

    override suspend fun updateCustomerById(
        customerId: String,
        customerUpdateDto: CustomerUpdateDto
    ): CustomerResponseDto {
        return customerService.updateCustomer(
            customerId = CustomerId(customerId),
            updateDto = customerUpdateDto
        )?.toResource() ?: throw NotFoundException()
    }

    override suspend fun getAllCustomers(limit: Int?, filter: String?): List<CustomerResponseDto> = when {
        filter?.trim().isNullOrBlank() -> customerRepository.findAll().map { it.toResource() }
        else -> searchCustomer(limit, filter)
    }

    override suspend fun getCustomerById(customerId: String): CustomerResponseDto {
        return customerRepository.findById(CustomerId(customerId))?.toResource() ?: throw NotFoundException()
    }

    override suspend fun getCustomerByPage(direction: String, pageToken: String?, limit: Int?): CustomersPaginatedDto {
        TODO("Not yet implemented")
    }

    override suspend fun searchCustomer(limit: Int?, filter: String?): List<CustomerResponseDto> {
        return customerRepository.findByKeyword(filter ?: "", limit ?: 10).map {
            it.toResource()
        }
    }

    object ResourceMapper {
        fun Customer.toResource() = CustomerResponseDto(
            id = this.id.value,
            name = this.name,
            surname = this.surname,
            phone = this.contacts.phone?.value,
            email = this.contacts.email?.value,
            note = this.note
        )
    }

}
