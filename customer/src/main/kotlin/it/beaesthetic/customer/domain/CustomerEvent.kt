package it.beaesthetic.customer.domain

sealed interface CustomerEvent

data class CustomerCreated(
    val customerId: CustomerId,
    val name: String,
    val surname: String?,
    val contacts: Contacts,
) : CustomerEvent

data class CustomerChanged(
    val customerId: CustomerId,
    val name: String,
    val surname: String?,
    val contacts: Contacts,
) : CustomerEvent
