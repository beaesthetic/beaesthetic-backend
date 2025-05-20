package it.beaesthetic.wallet.domain

interface WalletRepository {
    suspend fun findAll(): List<Wallet>

    suspend fun findByCustomer(owner: String): Wallet?

    suspend fun findById(id: WalletId): Wallet?

    suspend fun save(wallet: Wallet): Result<Wallet>
}
