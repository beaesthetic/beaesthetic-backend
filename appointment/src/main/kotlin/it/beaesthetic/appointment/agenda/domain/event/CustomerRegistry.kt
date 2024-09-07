package it.beaesthetic.appointment.agenda.domain.event

data class Customer(val customerId: String, val displayName: String, val phoneNumber: String?)

interface CustomerRegistry {
    suspend fun findByCustomerId(customerId: String): Customer?
}
