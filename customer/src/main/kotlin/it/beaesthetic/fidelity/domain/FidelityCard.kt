package it.beaesthetic.fidelity.domain

import java.util.UUID

@JvmInline
value class CustomerId(val id: String)

class FidelityCard(
    val id: String,
    val customerId: CustomerId,
    private var vouchers: List<Voucher> = emptyList(),
    private var purchaseCounter: Map<String, Int> = emptyMap()
) {

    companion object {
        fun ofCustomer(customerId: CustomerId) = FidelityCard(UUID.randomUUID().toString(), customerId)
    }

    val availableVouchers: List<Voucher> get() = vouchers

    fun purchasesOf(treatment: FidelityTreatment): Int =
        purchaseCounter.getOrDefault(treatment.name, 0)

    fun addPurchase(purchase: Purchase): Voucher? = when (purchase) {
        is TreatmentPurchase -> {
            val purchases = purchaseCounter.getOrDefault(purchase.treatment.name, 0)
            if (purchases + 1 >= 10) {
                vouchers += Voucher.ofTreatment(purchase.treatment)
                purchaseCounter += (purchase.treatment.name to 0)
                vouchers.last()
            } else {
                purchaseCounter += (purchase.treatment.name to purchases + 1)
                null
            }
        }
    }

    fun useVoucher(voucher: Voucher): Result<FidelityCard> {
        if (!voucher.isUsed) {
            return useVoucher(voucher.id)
        }
        return Result.failure(IllegalArgumentException("Voucher already used"))
    }

    fun useVoucher(voucherId: VoucherId): Result<FidelityCard> {
        val foundVoucher = vouchers.firstOrNull { it.id == voucherId }
        return when {
            foundVoucher == null -> Result.failure(IllegalStateException("No voucher found"))
            foundVoucher.isUsed -> Result.failure(IllegalStateException("Voucher with id $voucherId is already used"))
            else -> Result.success(
                FidelityCard(
                    id = id,
                    customerId = customerId,
                    vouchers = vouchers - foundVoucher + foundVoucher.copy(isUsed = true)
                )
            )
        }
    }
}