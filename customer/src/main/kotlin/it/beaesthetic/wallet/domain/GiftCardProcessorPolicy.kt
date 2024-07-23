package it.beaesthetic.wallet.domain

import java.time.Instant

class GiftCardProcessorPolicy {

    fun charge(
        giftCards: List<GiftCard>,
        amount: Money,
        now: Instant
    ): Pair<Money, List<GiftCard>> {
        val availableCards =
            giftCards.filter { !it.isExpiredAt(now) }.filter { it.availableAmount > Money.Zero }

        if (amount > Money.Zero && availableCards.isEmpty()) {
            return amount to giftCards
        }

        if (amount > Money.Zero) {
            val pickedCard = availableCards.first()
            val pickedCardAvailableAmount = pickedCard.availableAmount
            val updatedCard = pickedCard.partialCharge(amount, now).getOrThrow()
            return charge(
                availableCards - pickedCard + updatedCard,
                maxOf(Money.Zero, amount - pickedCardAvailableAmount),
                now
            )
        }
        return Money.Zero to giftCards
    }
}
