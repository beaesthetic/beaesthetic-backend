package it.beaesthetic.customer.application

import it.beaesthetic.customer.domain.Customer

interface CustomerReadRepository {
    suspend fun findNextPage(pageToken: String?, limit: Int?): CustomerPage
    data class CustomerPage(
        val customers: List<Customer>,
        val pageSize: Int,
        val nextToken: String?
    )
}
