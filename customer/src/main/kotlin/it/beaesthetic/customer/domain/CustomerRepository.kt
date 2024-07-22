package it.beaesthetic.customer.domain

interface CustomerRepository {
    suspend fun findByKeyword(keyword: String, maxResults: Int): List<Customer>
    suspend fun findById(id: CustomerId): Customer?
    suspend fun findAll(): List<Customer>
    suspend fun save(customer: Customer): Customer
}
