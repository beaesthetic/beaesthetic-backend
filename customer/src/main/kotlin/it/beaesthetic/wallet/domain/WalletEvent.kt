package it.beaesthetic.wallet.domain

import java.time.Instant

sealed interface WalletEvent {
    val at: Instant
}

data class MoneyCredited(
    override val at: Instant,
    val amount: Money,
) : WalletEvent

data class GiftCardMoneyCredited(
    override val at: Instant,
    val giftCardId: String,
    val amount: Money,
    val expiresAt: Instant,
) : WalletEvent

data class GiftCardMoneyExpired(
    override val at: Instant,
    val giftCardId: String,
) : WalletEvent

data class MoneyCharge(
    override val at: Instant,
    val amount: Money,
) : WalletEvent
