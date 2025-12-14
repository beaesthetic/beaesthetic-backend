package it.beaesthetic.customer.domain

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

class CustomerTest {

    private val name = "Mario"
    private val surname = "Rossi"
    private val email = Email("mario.rossi@example.com")
    private val phone = Phone.of("+393331234567")
    private val contacts = Contacts(email, phone)
    private val note = "VIP customer"

    @Test
    fun `should create Customer with generated ID`() {
        val customer = Customer.create(name, surname, contacts, note)

        assertNotNull(customer.id)
        assertEquals(name, customer.name)
        assertEquals(surname, customer.surname)
        assertEquals(contacts, customer.contacts)
        assertEquals(note, customer.note)
    }

    @Test
    fun `should emit CustomerCreated event when customer is created`() {
        val customer = Customer.create(name, surname, contacts, note)

        val events = customer.events
        assertEquals(1, events.size)

        val event = events.first().second as CustomerCreated
        assertEquals(customer.id, event.customerId)
        assertEquals(name, event.name)
        assertEquals(surname, event.surname)
        assertEquals(contacts, event.contacts)
    }

    @Test
    fun `should change contacts and emit CustomerChanged event`() {
        val customer = Customer.create(name, surname, contacts, note)
        customer.clearEvents()

        val newEmail = Email("new.email@example.com")
        val newPhone = Phone.of("+393339876543")
        val newContacts = Contacts(newEmail, newPhone)

        val updated = customer.changeContacts(newContacts)

        assertEquals(newContacts, updated.contacts)
        assertEquals(name, updated.name)
        assertEquals(surname, updated.surname)

        val events = updated.events
        assertEquals(1, events.size)

        val event = events.first().second as CustomerChanged
        assertEquals(customer.id, event.customerId)
        assertEquals(newContacts, event.contacts)
    }

    @Test
    fun `should update note without emitting events`() {
        val customer = Customer.create(name, surname, contacts, note)
        customer.clearEvents()

        val newNote = "Regular customer"
        val updated = customer.updateNote(newNote)

        assertEquals(newNote, updated.note)
        assertTrue(updated.events.isEmpty())
    }

    @Test
    fun `should preserve other fields when changing contacts`() {
        val customer = Customer.create(name, surname, contacts, note)
        val newContacts = Contacts(Email("other@example.com"), null)

        val updated = customer.changeContacts(newContacts)

        assertEquals(customer.id, updated.id)
        assertEquals(customer.name, updated.name)
        assertEquals(customer.surname, updated.surname)
        assertEquals(customer.note, updated.note)
    }

    @Test
    fun `should preserve other fields when updating note`() {
        val customer = Customer.create(name, surname, contacts, note)
        val newNote = "Updated note"

        val updated = customer.updateNote(newNote)

        assertEquals(customer.id, updated.id)
        assertEquals(customer.name, updated.name)
        assertEquals(customer.surname, updated.surname)
        assertEquals(customer.contacts, updated.contacts)
    }

    @Test
    fun `should handle contacts with only email`() {
        val contactsEmailOnly = Contacts(email, null)

        val customer = Customer.create(name, surname, contactsEmailOnly, note)

        assertEquals(contactsEmailOnly, customer.contacts)
        assertNotNull(customer.contacts.email)
        assertEquals(null, customer.contacts.phone)
    }

    @Test
    fun `should handle contacts with only phone`() {
        val contactsPhoneOnly = Contacts(null, phone)

        val customer = Customer.create(name, surname, contactsPhoneOnly, note)

        assertEquals(contactsPhoneOnly, customer.contacts)
        assertEquals(null, customer.contacts.email)
        assertNotNull(customer.contacts.phone)
    }

    @Test
    fun `should handle empty note`() {
        val customer = Customer.create(name, surname, contacts, "")

        assertEquals("", customer.note)
    }

    @Test
    fun `should generate different IDs for different customers`() {
        val customer1 = Customer.create(name, surname, contacts, note)
        val customer2 = Customer.create(name, surname, contacts, note)

        assertTrue(customer1.id.value != customer2.id.value)
    }
}
