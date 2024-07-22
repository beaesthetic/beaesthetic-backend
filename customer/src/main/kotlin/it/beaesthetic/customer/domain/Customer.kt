package it.beaesthetic.customer.domain

import java.util.UUID

@JvmInline
value class CustomerId(val value: String) {
    companion object {
        fun generate() = CustomerId(UUID.randomUUID().toString())
    }
}

data class Customer(
    val id: CustomerId,
    val name: String,
    val surname: String,
    val contacts: Contacts,
    val note: String,
    private val version: Int = 0
) {

    fun changeContacts(contacts: Contacts) = copy(contacts = contacts)

    fun updateNote(note: String) = copy(note = note)
}
