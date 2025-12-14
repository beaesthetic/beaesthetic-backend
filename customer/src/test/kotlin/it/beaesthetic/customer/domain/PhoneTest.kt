package it.beaesthetic.customer.domain

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class PhoneTest {

    @Test
    fun `should create Phone from valid full number with plus prefix`() {
        val fullNumber = "+393331234567"

        val phone = Phone.of(fullNumber)

        assertEquals("+39", phone.prefix)
        assertEquals("3331234567", phone.number)
        assertEquals("+393331234567", phone.fullNumber)
    }

    @Test
    fun `should create Phone from valid full number without plus prefix`() {
        val fullNumber = "393331234567"

        val phone = Phone.of(fullNumber)

        assertEquals("39", phone.prefix)
        assertEquals("3331234567", phone.number)
        assertEquals("393331234567", phone.fullNumber)
    }

    @Test
    fun `should fail when phone number has invalid format - too short`() {
        val invalidNumber = "+1"

        val exception = assertFailsWith<IllegalArgumentException> { Phone.of(invalidNumber) }

        assertEquals("Invalid phone number format: +1", exception.message)
    }

    @Test
    fun `should fail when phone number has invalid format - no digits`() {
        val invalidNumber = "abcdef"

        val exception = assertFailsWith<IllegalArgumentException> { Phone.of(invalidNumber) }

        assertEquals("Invalid phone number format: abcdef", exception.message)
    }

    @Test
    fun `should fail when phone number starts with letters`() {
        val invalidNumber = "abc123456789"

        val exception = assertFailsWith<IllegalArgumentException> { Phone.of(invalidNumber) }

        assertEquals("Invalid phone number format: abc123456789", exception.message)
    }

    @Test
    fun `should construct fullNumber correctly from prefix and number`() {
        val phone = Phone("+39", "3331234567")

        assertEquals("+393331234567", phone.fullNumber)
    }
}
