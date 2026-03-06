package it.beaesthetic.wallet.domain

import java.time.Duration
import java.time.Instant
import java.util.UUID

data class GiftCardVoucher(
    val id: String,
    val code: String,
    val amount: Money,
    val createdAt: Instant,
    val expiresAt: Instant,
) {
    companion object {
        private val ONE_YEAR = Duration.ofDays(365)

        fun issue(amount: Money, now: Instant = Instant.now()): GiftCardVoucher =
            GiftCardVoucher(
                id = UUID.randomUUID().toString(),
                code = generateCode(),
                amount = amount,
                createdAt = now,
                expiresAt = now + ONE_YEAR,
            )

        private fun generateCode(): String = (10000000..99999999).random().toString()
    }

    fun isExpiredAt(now: Instant): Boolean = now > expiresAt
}
