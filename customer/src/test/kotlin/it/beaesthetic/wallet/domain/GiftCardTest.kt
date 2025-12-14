package it.beaesthetic.wallet.domain

import java.time.Duration
import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class GiftCardTest {

    private val now = Instant.parse("2024-01-15T10:00:00Z")
    private val owner = "customer-123"
    private val id = "gift-card-1"

    @Test
    fun `should create GiftCard with factory method`() {
        val amount = Money(100.0)
        val expire = Duration.ofDays(365)

        val giftCard = GiftCard.of(id, owner, amount, now, expire)

        assertEquals(id, giftCard.id)
        assertEquals(owner, giftCard.owner)
        assertEquals(amount, giftCard.availableAmount)
        assertEquals(now, giftCard.createdAt)
        assertEquals(now.plus(expire), giftCard.expiresAt)
        assertEquals(Money.Zero, giftCard.amountSpent)
    }

    @Test
    fun `should not be expired when checked before expiration date`() {
        val giftCard = GiftCard.of(id, owner, Money(100.0), now, Duration.ofDays(30))
        val checkTime = now.plus(Duration.ofDays(15))

        val isExpired = giftCard.isExpiredAt(checkTime)

        assertFalse(isExpired)
    }

    @Test
    fun `should be expired when checked after expiration date`() {
        val giftCard = GiftCard.of(id, owner, Money(100.0), now, Duration.ofDays(30))
        val checkTime = now.plus(Duration.ofDays(31))

        val isExpired = giftCard.isExpiredAt(checkTime)

        assertTrue(isExpired)
    }

    @Test
    fun `should not be expired exactly at expiration time`() {
        val giftCard = GiftCard.of(id, owner, Money(100.0), now, Duration.ofDays(30))
        val checkTime = giftCard.expiresAt

        val isExpired = giftCard.isExpiredAt(checkTime)

        assertFalse(isExpired)
    }

    @Test
    fun `should charge full amount when available amount is sufficient`() {
        val giftCard = GiftCard.of(id, owner, Money(100.0), now, Duration.ofDays(30))
        val toCharge = Money(100.0)

        val result = giftCard.chargeMoney(toCharge, now)

        assertTrue(result.isSuccess)
        val charged = result.getOrThrow()
        assertEquals(Money(100.0), charged.amountSpent)
        assertEquals(Money(100.0), charged.availableAmount)
    }

    @Test
    fun `should fail to charge when amount is greater than available`() {
        val giftCard = GiftCard.of(id, owner, Money(50.0), now, Duration.ofDays(30))
        val toCharge = Money(100.0)

        val result = giftCard.chargeMoney(toCharge, now)

        assertTrue(result.isFailure)
        assertEquals("Not enough amount", result.exceptionOrNull()?.message)
    }

    @Test
    fun `should fail to charge when gift card is expired`() {
        val giftCard = GiftCard.of(id, owner, Money(100.0), now, Duration.ofDays(30))
        val expiredTime = now.plus(Duration.ofDays(31))
        val toCharge = Money(100.0)

        val result = giftCard.chargeMoney(toCharge, expiredTime)

        assertTrue(result.isFailure)
        assertEquals("GiftCard is expired", result.exceptionOrNull()?.message)
    }

    @Test
    fun `should perform partial charge when requested amount is greater than available`() {
        val giftCard = GiftCard.of(id, owner, Money(50.0), now, Duration.ofDays(30))
        val toCharge = Money(100.0)

        val result = giftCard.partialCharge(toCharge, now)

        assertTrue(result.isSuccess)
        val charged = result.getOrThrow()
        assertEquals(Money(50.0), charged.amountSpent)
        assertEquals(Money.Zero, charged.availableAmount)
    }

    @Test
    fun `should perform partial charge for exact available amount`() {
        val giftCard = GiftCard.of(id, owner, Money(75.0), now, Duration.ofDays(30))
        val toCharge = Money(75.0)

        val result = giftCard.partialCharge(toCharge, now)

        assertTrue(result.isSuccess)
        val charged = result.getOrThrow()
        assertEquals(Money(75.0), charged.amountSpent)
        assertEquals(Money.Zero, charged.availableAmount)
    }

    @Test
    fun `should perform partial charge for less than available amount`() {
        val giftCard = GiftCard.of(id, owner, Money(100.0), now, Duration.ofDays(30))
        val toCharge = Money(30.0)

        val result = giftCard.partialCharge(toCharge, now)

        assertTrue(result.isSuccess)
        val charged = result.getOrThrow()
        assertEquals(Money(30.0), charged.amountSpent)
        assertEquals(Money(70.0), charged.availableAmount)
    }

    @Test
    fun `should fail partial charge when gift card is expired`() {
        val giftCard = GiftCard.of(id, owner, Money(100.0), now, Duration.ofDays(30))
        val expiredTime = now.plus(Duration.ofDays(31))
        val toCharge = Money(50.0)

        val result = giftCard.partialCharge(toCharge, expiredTime)

        assertTrue(result.isFailure)
        assertEquals("GiftCard is expired", result.exceptionOrNull()?.message)
    }

    @Test
    fun `should handle multiple charges correctly`() {
        var giftCard = GiftCard.of(id, owner, Money(100.0), now, Duration.ofDays(30))

        // First charge
        giftCard = giftCard.partialCharge(Money(30.0), now).getOrThrow()
        assertEquals(Money(30.0), giftCard.amountSpent)
        assertEquals(Money(70.0), giftCard.availableAmount)

        // Second charge
        giftCard = giftCard.partialCharge(Money(50.0), now).getOrThrow()
        assertEquals(Money(80.0), giftCard.amountSpent)
        assertEquals(Money(20.0), giftCard.availableAmount)

        // Third charge (more than available)
        giftCard = giftCard.partialCharge(Money(50.0), now).getOrThrow()
        assertEquals(Money(100.0), giftCard.amountSpent)
        assertEquals(Money.Zero, giftCard.availableAmount)
    }
}
