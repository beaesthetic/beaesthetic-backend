package it.beaesthetic.wallet.domain

import java.time.Instant
import java.util.*

@JvmInline value class WalletId(val value: String)

data class Wallet(
    val id: String,
    val owner: String,
    val availableAmount: Money,
    val spentAmount: Money,
    val operations: List<WalletEvent>,
    val giftCards: List<GiftCard>,
    val createdAt: Instant = Instant.now(),
    val updatedAt: Instant = Instant.now(),
    private val processorPolicy: GiftCardProcessorPolicy,
) {

    companion object {
        fun ofCustomer(customer: String): Wallet =
            Wallet(
                id = UUID.randomUUID().toString(),
                owner = customer,
                availableAmount = Money.Zero,
                spentAmount = Money.Zero,
                operations = listOf(),
                giftCards = listOf(),
                processorPolicy = GiftCardProcessorPolicy(),
            )
    }

    fun charge(amount: Money, now: Instant = Instant.now()): Wallet {
        require(amount > Money.Zero) { "Amount must be positive" }
        require(amount <= availableAmount) {
            "Cannot redeem $amount cause exceed maximum amount of $availableAmount"
        }
        val updatedWallet = removeExpiredGiftCard(now)
        val (residual, giftCards) = processorPolicy.charge(updatedWallet.giftCards, amount, now)
        require(residual == Money.Zero) { "Cannot redeem $amount from gift cards" }
        return copy(
            availableAmount = updatedWallet.availableAmount - amount,
            operations = listOf(MoneyCharge(now, amount)) + operations,
            spentAmount = updatedWallet.spentAmount + amount,
            giftCards = giftCards,
        )
    }

    fun creditMoney(amount: Money, now: Instant = Instant.now()): Wallet {
        return copy(
            availableAmount = availableAmount + amount,
            operations = listOf(MoneyCredited(now, amount)) + operations,
        )
    }

    fun creditGiftCard(giftCard: GiftCard, now: Instant = Instant.now()): Wallet {
        require(giftCard.owner == owner) { "Gift card owner is not the same as the owner" }
        require(giftCards.none { it.id == giftCard.id }) { "Gift card already redeemed" }
        require(!giftCard.isExpiredAt(now)) { "Expired gift card cannot be redeemed" }
        return copy(
            availableAmount = availableAmount + giftCard.availableAmount,
            operations =
                listOf(
                    GiftCardMoneyCredited(
                        at = now,
                        giftCardId = giftCard.id,
                        expiresAt = giftCard.expiresAt,
                        amount = giftCard.availableAmount,
                    )
                ) + operations,
            giftCards = listOf(giftCard) + giftCards,
        )
    }

    fun expireGiftCard(giftCard: GiftCard, now: Instant = Instant.now()): Wallet {
        val expiredGiftCard = giftCards.find { it.id == giftCard.id }
        if (expiredGiftCard != null) {
            return copy(
                availableAmount = availableAmount - expiredGiftCard.availableAmount,
                operations =
                    listOf(GiftCardMoneyExpired(at = now, giftCardId = giftCard.id)) + operations,
                giftCards = giftCards - expiredGiftCard,
            )
        }
        return this
    }

    private fun removeExpiredGiftCard(now: Instant = Instant.now()): Wallet {
        return giftCards
            .filter { it.isExpiredAt(now) }
            .fold(this) { wallet, card -> wallet.expireGiftCard(card) }
    }
}
