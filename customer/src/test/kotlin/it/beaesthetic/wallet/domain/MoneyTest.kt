package it.beaesthetic.wallet.domain

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class MoneyTest {

    @Test
    fun `should create Money with positive amount`() {
        val money = Money(100.0)

        assertEquals(100.0, money.amount)
    }

    @Test
    fun `should create Money with zero amount`() {
        val money = Money.Zero

        assertEquals(0.0, money.amount)
    }

    @Test
    fun `should add two Money instances`() {
        val money1 = Money(50.0)
        val money2 = Money(30.0)

        val result = money1 + money2

        assertEquals(80.0, result.amount)
    }

    @Test
    fun `should subtract two Money instances`() {
        val money1 = Money(100.0)
        val money2 = Money(30.0)

        val result = money1 - money2

        assertEquals(70.0, result.amount)
    }

    @Test
    fun `should compare Money instances correctly`() {
        val smaller = Money(50.0)
        val larger = Money(100.0)
        val equal = Money(50.0)

        assertTrue(smaller < larger)
        assertTrue(larger > smaller)
        assertTrue(smaller == equal)
        assertTrue(smaller <= equal)
        assertTrue(smaller >= equal)
    }

    @Test
    fun `should handle addition with zero`() {
        val money = Money(50.0)

        val result = money + Money.Zero

        assertEquals(50.0, result.amount)
    }

    @Test
    fun `should handle subtraction resulting in zero`() {
        val money = Money(50.0)

        val result = money - Money(50.0)

        assertEquals(0.0, result.amount)
    }

    @Test
    fun `should support negative amounts from subtraction`() {
        val money1 = Money(30.0)
        val money2 = Money(50.0)

        val result = money1 - money2

        assertEquals(-20.0, result.amount)
    }

    @Test
    fun `should handle decimal precision in operations`() {
        val money1 = Money(10.50)
        val money2 = Money(5.25)

        val sum = money1 + money2
        val difference = money1 - money2

        assertEquals(15.75, sum.amount)
        assertEquals(5.25, difference.amount)
    }
}
