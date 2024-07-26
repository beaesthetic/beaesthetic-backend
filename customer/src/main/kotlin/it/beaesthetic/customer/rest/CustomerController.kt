package it.beaesthetic.customer.rest

import io.quarkus.cache.CacheInvalidate
import io.quarkus.cache.CacheKey
import io.quarkus.cache.CacheResult
import io.smallrye.mutiny.Uni
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.common.uniWithScope
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

    override fun createCustomer(
        customerCreateDto: CustomerCreateDto
    ): Uni<CreateCustomer201ResponseDto> = uniWithScope {
        CreateCustomer201ResponseDto(
            id = customerService.createCustomer(customerCreateDto).id.value
        )
    }

    @CacheInvalidate(cacheName = "customers")
    override fun updateCustomerById(
        @CacheKey customerId: String,
        customerUpdateDto: CustomerUpdateDto
    ): Uni<CustomerResponseDto> = uniWithScope {
        customerService
            .updateCustomer(customerId = CustomerId(customerId), updateDto = customerUpdateDto)
            ?.toResource()
            ?: throw NotFoundException()
    }

    @CacheResult(cacheName = "customers")
    override fun getCustomerById(@CacheKey customerId: String): Uni<CustomerResponseDto> =
        uniWithScope {
            customerRepository.findById(CustomerId(customerId))?.toResource()
                ?: throw NotFoundException()
        }

    @CacheResult(cacheName = "customers-search")
    override fun getAllCustomers(
        @CacheKey limit: Int?,
        @CacheKey filter: String?
    ): Uni<List<CustomerResponseDto>> = uniWithScope {
        when {
            filter?.trim().isNullOrBlank() -> customerRepository.findAll().map { it.toResource() }
            else -> searchCustomer(limit, filter).awaitSuspending()
        }
    }

    @CacheResult(cacheName = "customers-search")
    override fun searchCustomer(
        @CacheKey limit: Int?,
        @CacheKey filter: String?
    ): Uni<List<CustomerResponseDto>> = uniWithScope {
        customerRepository.findByKeyword(filter ?: "", limit ?: 10).map { it.toResource() }
    }

    override fun getCustomerByPage(
        direction: String,
        pageToken: String?,
        limit: Int?
    ): Uni<CustomersPaginatedDto> {
        TODO("Not yet implemented")
    }

    object ResourceMapper {
        fun Customer.toResource() =
            CustomerResponseDto(
                id = this.id.value,
                name = this.name,
                surname = this.surname,
                phone = this.contacts.phone?.fullNumber,
                email = this.contacts.email?.value,
                note = this.note
            )
    }
}
