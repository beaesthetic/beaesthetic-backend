package it.beaesthetic.customer.domain

import it.beaesthetic.common.DomainEventRegistry
import it.beaesthetic.common.DomainEventRegistryDelegate
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
    val domainEventRegistry: DomainEventRegistry<CustomerEvent>,
    private val version: Int = 0,
) : DomainEventRegistry<CustomerEvent> by domainEventRegistry {

    companion object {
        fun create(name: String, surname: String, contacts: Contacts, note: String) =
            Customer(
                    id = CustomerId.generate(),
                    name = name,
                    surname = surname,
                    contacts = contacts,
                    note = note,
                    domainEventRegistry = DomainEventRegistryDelegate(),
                )
                .apply { addEvent("CustomerCreated", CustomerCreated(id, name, surname, contacts)) }
    }

    fun changeContacts(contacts: Contacts): Customer {
        domainEventRegistry.addEvent(
            "CustomerChanged",
            CustomerChanged(id, name, surname, contacts),
        )
        return copy(contacts = contacts)
    }

    fun updateNote(note: String) = copy(note = note)
}
