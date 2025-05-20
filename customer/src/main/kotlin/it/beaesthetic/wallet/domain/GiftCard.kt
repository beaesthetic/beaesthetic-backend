package it.beaesthetic.wallet.domain

import java.time.Duration
import java.time.Instant

data class GiftCard(
    val id: String,
    val owner: String,
    val availableAmount: Money,
    val createdAt: Instant,
    val expiresAt: Instant,
    val amountSpent: Money,
) {

    companion object {
        fun of(
            id: String,
            owner: String,
            amount: Money,
            createdAt: Instant,
            expire: Duration,
        ): GiftCard {
            return GiftCard(id, owner, amount, createdAt, createdAt + expire, Money.Zero)
        }
    }

    fun isExpiredAt(now: Instant): Boolean = now > expiresAt

    fun chargeMoney(money: Money, now: Instant?): Result<GiftCard> = runCatching {
        assertMoneyAndExpire(money, now)
        copy(amountSpent = amountSpent + money)
    }

    fun partialCharge(money: Money, now: Instant?): Result<GiftCard> = runCatching {
        val toSpent = minOf(availableAmount, money)
        require(!isExpiredAt(now ?: Instant.now())) { "GiftCard is expired" }
        copy(amountSpent = amountSpent + toSpent, availableAmount = availableAmount - toSpent)
    }

    private fun assertMoneyAndExpire(money: Money, now: Instant?) {
        require(availableAmount <= money) { "Amount must be greater than zero" }
        require(!isExpiredAt(now ?: Instant.now())) { "GiftCard is expired" }
    }
}
