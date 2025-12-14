package it.beaesthetic.wallet.domain

import java.time.Duration
import java.time.Instant
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertTrue

class WalletTest {

    private val now = Instant.parse("2024-01-15T10:00:00Z")
    private val customerId = "customer-123"

    @Test
    fun `should create Wallet for customer with zero balance`() {
        val wallet = Wallet.ofCustomer(customerId)

        assertEquals(customerId, wallet.owner)
        assertEquals(Money.Zero, wallet.availableAmount)
        assertEquals(Money.Zero, wallet.spentAmount)
        assertTrue(wallet.operations.isEmpty())
        assertTrue(wallet.giftCards.isEmpty())
    }

    @Test
    fun `should credit money to wallet`() {
        val wallet = Wallet.ofCustomer(customerId)
        val amount = Money(100.0)

        val updated = wallet.creditMoney(amount, now)

        assertEquals(Money(100.0), updated.availableAmount)
        assertEquals(1, updated.operations.size)

        val operation = updated.operations.first() as MoneyCredited
        assertEquals(now, operation.at)
        assertEquals(amount, operation.amount)
    }

    @Test
    fun `should credit multiple amounts to wallet`() {
        var wallet = Wallet.ofCustomer(customerId)

        wallet = wallet.creditMoney(Money(50.0), now)
        wallet = wallet.creditMoney(Money(30.0), now)

        assertEquals(Money(80.0), wallet.availableAmount)
        assertEquals(2, wallet.operations.size)
    }

    @Test
    fun `should credit gift card to wallet`() {
        val wallet = Wallet.ofCustomer(customerId)
        val giftCard = GiftCard.of("gc-1", customerId, Money(100.0), now, Duration.ofDays(30))

        val updated = wallet.creditGiftCard(giftCard, now)

        assertEquals(Money(100.0), updated.availableAmount)
        assertEquals(1, updated.giftCards.size)
        assertEquals(giftCard, updated.giftCards.first())

        val operation = updated.operations.first() as GiftCardMoneyCredited
        assertEquals(now, operation.at)
        assertEquals("gc-1", operation.giftCardId)
        assertEquals(Money(100.0), operation.amount)
    }

    @Test
    fun `should fail to credit gift card with different owner`() {
        val wallet = Wallet.ofCustomer(customerId)
        val giftCard =
            GiftCard.of("gc-1", "different-customer", Money(100.0), now, Duration.ofDays(30))

        val exception =
            assertFailsWith<IllegalArgumentException> { wallet.creditGiftCard(giftCard, now) }

        assertEquals("Gift card owner is not the same as the owner", exception.message)
    }

    @Test
    fun `should fail to credit already redeemed gift card`() {
        val wallet = Wallet.ofCustomer(customerId)
        val giftCard = GiftCard.of("gc-1", customerId, Money(100.0), now, Duration.ofDays(30))

        val updated = wallet.creditGiftCard(giftCard, now)

        val exception =
            assertFailsWith<IllegalArgumentException> { updated.creditGiftCard(giftCard, now) }

        assertEquals("Gift card already redeemed", exception.message)
    }

    @Test
    fun `should fail to credit expired gift card`() {
        val wallet = Wallet.ofCustomer(customerId)
        val giftCard = GiftCard.of("gc-1", customerId, Money(100.0), now, Duration.ofDays(30))
        val futureTime = now.plus(Duration.ofDays(31))

        val exception =
            assertFailsWith<IllegalArgumentException> {
                wallet.creditGiftCard(giftCard, futureTime)
            }

        assertEquals("Expired gift card cannot be redeemed", exception.message)
    }

    @Test
    fun `should charge money from wallet with sufficient balance`() {
        var wallet = Wallet.ofCustomer(customerId)
        val giftCard = GiftCard.of("gc-1", customerId, Money(100.0), now, Duration.ofDays(30))
        wallet = wallet.creditGiftCard(giftCard, now)

        val charged = wallet.charge(Money(50.0), now)

        assertEquals(Money(50.0), charged.availableAmount)
        assertEquals(Money(50.0), charged.spentAmount)

        val operation = charged.operations.first() as MoneyCharge
        assertEquals(Money(50.0), operation.amount)
    }

    @Test
    fun `should fail to charge more than available amount`() {
        var wallet = Wallet.ofCustomer(customerId)
        val giftCard = GiftCard.of("gc-1", customerId, Money(100.0), now, Duration.ofDays(30))
        wallet = wallet.creditGiftCard(giftCard, now)

        val exception =
            assertFailsWith<IllegalArgumentException> { wallet.charge(Money(150.0), now) }

        assertTrue(exception.message!!.contains("Cannot redeem"))
    }

    @Test
    fun `should fail to charge zero or negative amount`() {
        var wallet = Wallet.ofCustomer(customerId)
        val giftCard = GiftCard.of("gc-1", customerId, Money(100.0), now, Duration.ofDays(30))
        wallet = wallet.creditGiftCard(giftCard, now)

        val exception = assertFailsWith<IllegalArgumentException> { wallet.charge(Money.Zero, now) }

        assertEquals("Amount must be positive", exception.message)
    }

    @Test
    fun `should expire gift card and remove from available amount`() {
        var wallet = Wallet.ofCustomer(customerId)
        val giftCard = GiftCard.of("gc-1", customerId, Money(100.0), now, Duration.ofDays(30))
        wallet = wallet.creditGiftCard(giftCard, now)

        val updated = wallet.expireGiftCard(giftCard, now)

        assertEquals(Money.Zero, updated.availableAmount)
        assertTrue(updated.giftCards.isEmpty())

        val operation = updated.operations.first() as GiftCardMoneyExpired
        assertEquals("gc-1", operation.giftCardId)
    }

    @Test
    fun `should not modify wallet when expiring non-existent gift card`() {
        val wallet = Wallet.ofCustomer(customerId)
        val giftCard = GiftCard.of("gc-1", customerId, Money(100.0), now, Duration.ofDays(30))

        val updated = wallet.expireGiftCard(giftCard, now)

        assertEquals(wallet, updated)
    }

    @Test
    fun `should automatically remove expired gift cards on charge`() {
        var wallet = Wallet.ofCustomer(customerId)

        // Add two gift cards
        val giftCard1 = GiftCard.of("gc-1", customerId, Money(50.0), now, Duration.ofDays(30))
        val giftCard2 = GiftCard.of("gc-2", customerId, Money(50.0), now, Duration.ofDays(60))
        wallet = wallet.creditGiftCard(giftCard1, now)
        wallet = wallet.creditGiftCard(giftCard2, now)

        assertEquals(Money(100.0), wallet.availableAmount)
        assertEquals(2, wallet.giftCards.size)

        // Charge after first card expires but second is still valid
        val futureTime = now.plus(Duration.ofDays(31))
        val charged = wallet.charge(Money(30.0), futureTime)

        // First gift card should be automatically expired
        assertEquals(1, charged.giftCards.size)
        assertEquals("gc-2", charged.giftCards.first().id)
        assertEquals(Money(20.0), charged.availableAmount) // 50 (gc-2) - 30 (charged)
    }

    @Test
    fun `should handle multiple gift cards and charge from them`() {
        var wallet = Wallet.ofCustomer(customerId)

        val giftCard1 = GiftCard.of("gc-1", customerId, Money(30.0), now, Duration.ofDays(30))
        val giftCard2 = GiftCard.of("gc-2", customerId, Money(70.0), now, Duration.ofDays(30))

        wallet = wallet.creditGiftCard(giftCard1, now)
        wallet = wallet.creditGiftCard(giftCard2, now)

        assertEquals(Money(100.0), wallet.availableAmount)

        val charged = wallet.charge(Money(80.0), now)

        assertEquals(Money(20.0), charged.availableAmount)
        assertEquals(Money(80.0), charged.spentAmount)
    }

    @Test
    fun `should preserve operations history when performing multiple operations`() {
        var wallet = Wallet.ofCustomer(customerId)

        wallet = wallet.creditMoney(Money(50.0), now)
        val giftCard = GiftCard.of("gc-1", customerId, Money(50.0), now, Duration.ofDays(30))
        wallet = wallet.creditGiftCard(giftCard, now)
        wallet = wallet.charge(Money(30.0), now)

        assertEquals(3, wallet.operations.size)
        assertTrue(wallet.operations[0] is MoneyCharge)
        assertTrue(wallet.operations[1] is GiftCardMoneyCredited)
        assertTrue(wallet.operations[2] is MoneyCredited)
    }
}
