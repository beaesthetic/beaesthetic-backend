package it.beaesthetic.appointment.agenda.domain

data class Customer(
    val customerId: String,
    val displayName: String,
)

interface CustomerRegistry {
    suspend fun findByCustomerId(customerId: String): Customer?
}