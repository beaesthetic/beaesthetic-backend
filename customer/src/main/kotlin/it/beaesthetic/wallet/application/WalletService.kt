package it.beaesthetic.wallet.application

import it.beaesthetic.wallet.domain.*
import java.time.Duration
import java.time.Instant

class WalletService(
    private val walletRepository: WalletRepository,
    private val voucherRepository: GiftCardVoucherRepository,
) {

    suspend fun createGiftCardVoucher(amount: Money): Result<GiftCardVoucher> {
        val pending = GiftCardVoucher.issue(amount)
        return voucherRepository.save(pending)
    }

    suspend fun redeemGiftCard(code: String, customerId: String): Result<Wallet> {
        val pending = voucherRepository.findByCode(code)
            ?: return Result.failure(IllegalArgumentException("Gift card code not found"))

        if (pending.isExpiredAt(Instant.now())) {
            return Result.failure(IllegalArgumentException("Gift card is expired"))
        }

        val wallet = walletRepository.findByCustomer(customerId) ?: Wallet.ofCustomer(customerId)
        val giftCard = GiftCard.of(
            id = pending.id,
            owner = customerId,
            amount = pending.amount,
            createdAt = pending.createdAt,
            expire = Duration.between(pending.createdAt, pending.expiresAt),
        )

        return walletRepository.save(wallet.creditGiftCard(giftCard, Instant.now()))
            .also { result -> if (result.isSuccess) voucherRepository.delete(pending.id) }
    }

    suspend fun getGiftCard(id: String): GiftCardVoucher? =
        voucherRepository.findById(id)

    suspend fun listGiftCards(limit: Int, offset: Int): List<GiftCardVoucher> =
        voucherRepository.findPage(limit, offset)

    suspend fun getGiftCardVouchers(): List<GiftCardVoucher> =
        voucherRepository.findAll()
            .filter { !it.isExpiredAt(Instant.now()) }

    suspend fun redeemGiftCardById(id: String, customerId: String): Result<Wallet> {
        val pending = voucherRepository.findById(id)
            ?: return Result.failure(IllegalArgumentException("Gift card not found"))
        return redeemGiftCard(pending.code, customerId)
    }

    suspend fun charge(walletId: WalletId, money: Money): Result<Wallet> {
        val wallet = walletRepository.findById(walletId)
        return wallet?.charge(money, Instant.now())?.let { walletRepository.save(it) }
            ?: Result.failure(IllegalArgumentException("Wallet $walletId not found"))
    }
}
