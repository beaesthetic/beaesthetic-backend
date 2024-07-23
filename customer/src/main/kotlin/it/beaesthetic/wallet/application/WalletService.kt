package it.beaesthetic.wallet.application

import it.beaesthetic.wallet.domain.*
import java.time.Duration
import java.time.Instant
import java.util.UUID

class WalletService(private val walletRepository: WalletRepository) {

    companion object {
        private val ONE_YEAR = Duration.ofDays(365)
    }

    suspend fun createWallet(customer: String, amount: Money): Result<Wallet> {
        val existingWallet =
            walletRepository.findByCustomer(customer) ?: Wallet.ofCustomer(customer)
        val giftCard =
            GiftCard.of(
                id = UUID.randomUUID().toString(),
                owner = customer,
                amount = amount,
                createdAt = Instant.now(),
                expire = ONE_YEAR,
            )
        return walletRepository.save(existingWallet.creditGiftCard(giftCard, Instant.now()))
    }

    suspend fun charge(walletId: WalletId, money: Money): Result<Wallet> {
        val wallet = walletRepository.findById(walletId)
        return wallet?.charge(money, Instant.now())?.let { walletRepository.save(it) }
            ?: Result.failure(IllegalArgumentException("Wallet $walletId not found"))
    }
}
