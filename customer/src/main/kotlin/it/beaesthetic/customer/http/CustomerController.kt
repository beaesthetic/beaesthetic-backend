package it.beaesthetic.customer.http

import io.quarkus.cache.Cache
import io.quarkus.cache.CacheName
import io.quarkus.panache.common.Sort
import it.beaesthetic.common.SuspendableCache
import it.beaesthetic.customer.application.CustomerReadRepository
import it.beaesthetic.customer.application.CustomerService
import it.beaesthetic.customer.domain.Customer
import it.beaesthetic.customer.domain.CustomerId
import it.beaesthetic.customer.domain.CustomerRepository
import it.beaesthetic.customer.generated.api.CustomersAdminApi
import it.beaesthetic.customer.generated.api.model.*
import it.beaesthetic.customer.http.CustomerController.ResourceMapper.toResource
import jakarta.ws.rs.ClientErrorException
import jakarta.ws.rs.NotFoundException
import jakarta.ws.rs.WebApplicationException
import jakarta.ws.rs.core.Response
import java.math.BigDecimal

class CustomerController(
    private val customerService: CustomerService,
    private val customerRepository: CustomerRepository,
    private val customerReadRepository: CustomerReadRepository,
    private val cacheWrapper: SuspendableCache,
    @CacheName("customers") private val customerCache: Cache,
    @CacheName("customers-search") private val customerSearchCache: Cache,
) : CustomersAdminApi {

    override suspend fun createCustomer(
        customerCreateDto: CustomerCreateDto
    ): CreateCustomer201ResponseDto =
        CreateCustomer201ResponseDto(
            id = customerService.createCustomer(customerCreateDto).id.value
        )

    override suspend fun deleteCustomer(customerId: String) {
        customerService.deleteCustomer(CustomerId(customerId))
        cacheWrapper.invalidateKey(customerCache, customerId)
    }

    override suspend fun updateCustomerById(
        customerId: String,
        customerUpdateDto: CustomerUpdateDto,
    ): CustomerResponseDto =
        cacheWrapper.getOrLoad(customerCache, customerId) {
            customerService
                .updateCustomer(customerId = CustomerId(customerId), updateDto = customerUpdateDto)
                ?.toResource() ?: throw NotFoundException()
        }

    override suspend fun getCustomerById(customerId: String): CustomerResponseDto =
        cacheWrapper.getOrLoad(customerCache, customerId) {
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
        val page =
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
        cacheWrapper.getOrLoad(customerSearchCache, limit, filter) {
            when {
                filter?.trim().isNullOrBlank() ->
                    customerRepository.findAll().map { it.toResource() }
                else -> searchCustomer(limit, filter)
            }
        }

    override suspend fun searchCustomer(limit: Int?, filter: String?): List<CustomerResponseDto> =
        cacheWrapper.getOrLoad(customerSearchCache, limit, filter) {
            customerRepository.findByKeyword(filter ?: "", limit ?: 10).map { it.toResource() }
        }

    override suspend fun searchCustomerByPhone(
        searchCustomerByPhoneRequestDto: SearchCustomerByPhoneRequestDto
    ): CustomerResponseDto {
        return customerReadRepository
            .findByPhoneNumber(searchCustomerByPhoneRequestDto.phone)
            ?.toResource() ?: throw ClientErrorException(Response.Status.NOT_FOUND)
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
