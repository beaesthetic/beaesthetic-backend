package it.beaesthetic.customer.rest

import io.quarkus.panache.common.Sort
import it.beaesthetic.common.SuspendableCache
import it.beaesthetic.customer.application.CustomerReadRepository
import it.beaesthetic.customer.application.CustomerService
import it.beaesthetic.customer.domain.Customer
import it.beaesthetic.customer.domain.CustomerId
import it.beaesthetic.customer.domain.CustomerRepository
import it.beaesthetic.customer.generated.api.CustomersApi
import it.beaesthetic.customer.generated.api.model.*
import it.beaesthetic.customer.rest.CustomerController.ResourceMapper.toResource
import jakarta.ws.rs.NotFoundException
import jakarta.ws.rs.WebApplicationException
import jakarta.ws.rs.core.Response
import java.math.BigDecimal

class CustomerController(
    private val customerService: CustomerService,
    private val customerRepository: CustomerRepository,
    private val customerReadRepository: CustomerReadRepository,
    private val cache: SuspendableCache,
) : CustomersApi {

    companion object {
        private const val CACHE_CUSTOMERS = "customers"

        private const val CACHE_CUSTOMERS_SEARCH = "customers-search"
    }

    override suspend fun createCustomer(
        customerCreateDto: CustomerCreateDto
    ): CreateCustomer201ResponseDto =
        CreateCustomer201ResponseDto(
            id = customerService.createCustomer(customerCreateDto).id.value
        )

    override suspend fun deleteCustomer(customerId: String) {
        customerService.deleteCustomer(CustomerId(customerId))
        cache.invalidateKey(CACHE_CUSTOMERS, customerId)
    }

    override suspend fun updateCustomerById(
        customerId: String,
        customerUpdateDto: CustomerUpdateDto,
    ): CustomerResponseDto =
        cache.getOrLoad(CACHE_CUSTOMERS, customerId) {
            customerService
                .updateCustomer(customerId = CustomerId(customerId), updateDto = customerUpdateDto)
                ?.toResource() ?: throw NotFoundException()
        }

    override suspend fun getCustomerById(customerId: String): CustomerResponseDto =
        cache.getOrLoad(CACHE_CUSTOMERS, customerId) {
            customerRepository.findById(CustomerId(customerId))?.toResource()
                ?: throw NotFoundException()
        }

    override suspend fun getCustomerByPage(
        direction: String,
        pageToken: String?,
        limit: Int?,
        sortBy: String?,
    ): CustomersPaginatedDto {
        if (direction == "prev") {
            throw WebApplicationException(
                "Backward pagination not yet implemented",
                Response.Status.NOT_IMPLEMENTED,
            )
        }
        var page =
            customerReadRepository.findNextPage(
                pageToken,
                limit,
                if (sortBy != null) listOf(sortBy) else listOf("name", "surname"),
                sortDirection = Sort.Direction.Ascending,
            )
        return CustomersPaginatedDto(
            BigDecimal(page.pageSize),
            page.customers.map { it.toResource() },
            nextCursor = page.nextToken,
            hasNextPage = page.nextToken != null,
            hasPreviousPage = false,
        )
    }

    override suspend fun getAllCustomers(limit: Int?, filter: String?): List<CustomerResponseDto> =
        cache.getOrLoad(CACHE_CUSTOMERS_SEARCH, limit, filter) {
            when {
                filter?.trim().isNullOrBlank() ->
                    customerRepository.findAll().map { it.toResource() }
                else -> searchCustomer(limit, filter)
            }
        }

    override suspend fun searchCustomer(limit: Int?, filter: String?): List<CustomerResponseDto> =
        cache.getOrLoad(CACHE_CUSTOMERS_SEARCH, limit, filter) {
            customerRepository.findByKeyword(filter ?: "", limit ?: 10).map { it.toResource() }
        }

    object ResourceMapper {
        fun Customer.toResource() =
            CustomerResponseDto(
                id = this.id.value,
                name = this.name,
                surname = this.surname,
                phone = this.contacts.phone?.fullNumber,
                email = this.contacts.email?.value,
                note = this.note,
            )
    }
}
