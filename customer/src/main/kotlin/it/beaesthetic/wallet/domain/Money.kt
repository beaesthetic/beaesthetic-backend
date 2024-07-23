package it.beaesthetic.wallet.domain

data class Money(
    val amount: Double,
) : Comparable<Money> {

    companion object {
        val Zero = Money(0.0)
    }

    operator fun plus(other: Money): Money = copy(amount = amount + other.amount)

    operator fun minus(other: Money): Money = copy(amount = amount - other.amount)

    override fun compareTo(other: Money): Int {
        return amount.compareTo(other.amount)
    }
}
