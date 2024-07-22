package it.beaesthetic.fidelity.application

import it.beaesthetic.fidelity.domain.*

class FidelityCardService(private val fidelityCardRepository: FidelityCardRepository) {

    suspend fun createFidelityCard(customerId: CustomerId): Result<FidelityCard> {
        if (fidelityCardRepository.findByCustomerId(customerId) != null) {
            return Result.failure(IllegalArgumentException("Fidelity card already exists"))
        }

        val fidelityCard = FidelityCard.ofCustomer(customerId)
        return fidelityCardRepository.save(fidelityCard)
    }

    suspend fun registerPurchase(cardId: String, purchase: Purchase): Result<Voucher?> {
        val fidelityCard = fidelityCardRepository.findById(cardId)
        if (fidelityCard == null) {
            return Result.failure(IllegalArgumentException("Fidelity card not found"))
        } else {
            val voucher = fidelityCard.addPurchase(purchase)
            return fidelityCardRepository.save(fidelityCard).map { voucher }
        }
    }

    suspend fun useVoucher(voucherId: VoucherId): Result<FidelityCard> {
        val fidelityCard =
            fidelityCardRepository.findByVoucherId(voucherId.value)?.useVoucher(voucherId)
        return fidelityCard ?: Result.failure(IllegalArgumentException("Fidelity card not found"))
    }
}
