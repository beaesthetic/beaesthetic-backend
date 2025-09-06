package it.beaesthetic.customer.application

import io.quarkus.panache.common.Sort
import it.beaesthetic.customer.domain.Customer

interface CustomerReadRepository {

    suspend fun findByPhoneNumber(phoneNumber: String): Customer?

    suspend fun findNextPage(
        pageToken: String?,
        limit: Int?,
        sortBy: List<String>,
        sortDirection: Sort.Direction,
    ): CustomerPage

    data class CustomerPage(
        val customers: List<Customer>,
        val pageSize: Int,
        val nextToken: String?,
    )
}
