package it.beaesthetic.customer.application

import it.beaesthetic.customer.domain.*
import it.beaesthetic.customer.generated.api.model.CustomerCreateDto
import it.beaesthetic.customer.generated.api.model.CustomerUpdateDto
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking
import org.mockito.kotlin.*

class CustomerServiceTest {

    private val customerRepository: CustomerRepository = mock()
    private val service = CustomerService(customerRepository)

    @Test
    fun `should create customer with all fields`(): Unit = runBlocking {
        // Given
        val createDto =
            CustomerCreateDto(
                name = "Mario",
                surname = "Rossi",
                email = "mario.rossi@example.com",
                phone = "+393331234567",
                note = "VIP customer",
            )

        whenever(customerRepository.save(any())).thenAnswer { invocation ->
            invocation.getArgument<Customer>(0)
        }

        // When
        val result = service.createCustomer(createDto)

        // Then
        assertEquals("Mario", result.name)
        assertEquals("Rossi", result.surname)
        assertEquals("mario.rossi@example.com", result.contacts.email?.value)
        assertEquals("+393331234567", result.contacts.phone?.fullNumber)
        assertEquals("VIP customer", result.note)

        verify(customerRepository).save(any())
    }

    @Test
    fun `should create customer with minimal fields`(): Unit = runBlocking {
        // Given
        val createDto =
            CustomerCreateDto(
                name = "Mario",
                surname = null,
                email = null,
                phone = null,
                note = null,
            )

        whenever(customerRepository.save(any())).thenAnswer { invocation ->
            invocation.getArgument<Customer>(0)
        }

        // When
        val result = service.createCustomer(createDto)

        // Then
        assertEquals("Mario", result.name)
        assertEquals("", result.surname)
        assertNull(result.contacts.email)
        assertNull(result.contacts.phone)
        assertEquals("", result.note)

        verify(customerRepository).save(any())
    }

    @Test
    fun `should emit CustomerCreated event when creating customer`(): Unit = runBlocking {
        // Given
        val createDto =
            CustomerCreateDto(
                name = "Mario",
                surname = "Rossi",
                email = "mario@example.com",
                phone = "+393331234567",
                note = "Test",
            )

        whenever(customerRepository.save(any())).thenAnswer { invocation ->
            invocation.getArgument<Customer>(0)
        }

        // When
        val result = service.createCustomer(createDto)

        // Then
        val events = result.events
        assertEquals(1, events.size)
        assertTrue(events.first().second is CustomerCreated)

        val event = events.first().second as CustomerCreated
        assertEquals(result.id, event.customerId)
        assertEquals("Mario", event.name)
    }

    @Test
    fun `should update customer with all fields`(): Unit = runBlocking {
        // Given
        val customerId = CustomerId.generate()
        val existingCustomer =
            Customer.create(
                    name = "Mario",
                    surname = "Rossi",
                    contacts = Contacts(Email("old@example.com"), Phone.of("+393331111111")),
                    note = "Old note",
                )
                .copy(id = customerId)

        val updateDto =
            CustomerUpdateDto(
                name = "Luigi",
                surname = "Verdi",
                email = "new@example.com",
                phone = "+393339999999",
                note = "New note",
            )

        whenever(customerRepository.findById(customerId)).thenReturn(existingCustomer)
        whenever(customerRepository.save(any())).thenAnswer { invocation ->
            invocation.getArgument<Customer>(0)
        }

        // When
        val result = service.updateCustomer(customerId, updateDto)

        // Then
        assertNotNull(result)
        assertEquals("Luigi", result.name)
        assertEquals("Verdi", result.surname)
        assertEquals("new@example.com", result.contacts.email?.value)
        assertEquals("+393339999999", result.contacts.phone?.fullNumber)
        assertEquals("New note", result.note)

        verify(customerRepository).findById(customerId)
        verify(customerRepository).save(any())
    }

    @Test
    fun `should update customer with partial fields`(): Unit = runBlocking {
        // Given
        val customerId = CustomerId.generate()
        val existingCustomer =
            Customer.create(
                    name = "Mario",
                    surname = "Rossi",
                    contacts = Contacts(Email("mario@example.com"), Phone.of("+393331111111")),
                    note = "Old note",
                )
                .copy(id = customerId)

        val updateDto =
            CustomerUpdateDto(
                name = "Luigi",
                surname = null,
                email = null,
                phone = null,
                note = null,
            )

        whenever(customerRepository.findById(customerId)).thenReturn(existingCustomer)
        whenever(customerRepository.save(any())).thenAnswer { invocation ->
            invocation.getArgument<Customer>(0)
        }

        // When
        val result = service.updateCustomer(customerId, updateDto)

        // Then
        assertNotNull(result)
        assertEquals("Luigi", result.name)
        assertEquals("Rossi", result.surname) // Unchanged
        assertEquals("mario@example.com", result.contacts.email?.value) // Unchanged
        assertEquals("+393331111111", result.contacts.phone?.fullNumber) // Unchanged
        assertEquals("Old note", result.note) // Unchanged

        verify(customerRepository).findById(customerId)
        verify(customerRepository).save(any())
    }

    @Test
    fun `should emit CustomerChanged event when updating contacts`(): Unit = runBlocking {
        // Given
        val customerId = CustomerId.generate()
        val existingCustomer =
            Customer.create(
                    name = "Mario",
                    surname = "Rossi",
                    contacts = Contacts(Email("old@example.com"), null),
                    note = "Note",
                )
                .copy(id = customerId)
        existingCustomer.clearEvents()

        val updateDto =
            CustomerUpdateDto(
                name = null,
                surname = null,
                email = "new@example.com",
                phone = null,
                note = null,
            )

        whenever(customerRepository.findById(customerId)).thenReturn(existingCustomer)
        whenever(customerRepository.save(any())).thenAnswer { invocation ->
            invocation.getArgument<Customer>(0)
        }

        // When
        val result = service.updateCustomer(customerId, updateDto)

        // Then
        assertNotNull(result)
        val events = result.events
        assertEquals(1, events.size)
        assertTrue(events.first().second is CustomerChanged)
    }

    @Test
    fun `should return null when updating non-existent customer`(): Unit = runBlocking {
        // Given
        val customerId = CustomerId.generate()
        val updateDto =
            CustomerUpdateDto(
                name = "Luigi",
                surname = null,
                email = null,
                phone = null,
                note = null,
            )

        whenever(customerRepository.findById(customerId)).thenReturn(null)

        // When
        val result = service.updateCustomer(customerId, updateDto)

        // Then
        assertNull(result)
        verify(customerRepository).findById(customerId)
        verify(customerRepository, never()).save(any())
    }

    @Test
    fun `should delete customer by ID`(): Unit = runBlocking {
        // Given
        val customerId = CustomerId.generate()

        whenever(customerRepository.delete(customerId)).thenReturn(true)

        // When
        val result = service.deleteCustomer(customerId)

        // Then
        assertTrue(result)
        verify(customerRepository).delete(customerId)
    }

    @Test
    fun `should return false when deleting non-existent customer`(): Unit = runBlocking {
        // Given
        val customerId = CustomerId.generate()

        whenever(customerRepository.delete(customerId)).thenReturn(false)

        // When
        val result = service.deleteCustomer(customerId)

        // Then
        assertTrue(!result)
        verify(customerRepository).delete(customerId)
    }

    @Test
    fun `should preserve customer ID when updating`(): Unit = runBlocking {
        // Given
        val customerId = CustomerId.generate()
        val existingCustomer =
            Customer.create(
                    name = "Mario",
                    surname = "Rossi",
                    contacts = Contacts(null, null),
                    note = "",
                )
                .copy(id = customerId)

        val updateDto =
            CustomerUpdateDto(
                name = "Luigi",
                surname = null,
                email = null,
                phone = null,
                note = null,
            )

        whenever(customerRepository.findById(customerId)).thenReturn(existingCustomer)
        whenever(customerRepository.save(any())).thenAnswer { invocation ->
            invocation.getArgument<Customer>(0)
        }

        // When
        val result = service.updateCustomer(customerId, updateDto)

        // Then
        assertNotNull(result)
        assertEquals(customerId, result.id)
    }
}
