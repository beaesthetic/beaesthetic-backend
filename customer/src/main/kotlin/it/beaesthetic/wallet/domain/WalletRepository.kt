package it.beaesthetic.wallet.domain

import it.beaesthetic.wallet.application.read.WalletReadModel

interface WalletRepository {
    suspend fun findAll(): List<WalletReadModel>

    suspend fun findByCustomer(owner: String): Wallet?

    suspend fun findById(id: WalletId): Wallet?

    suspend fun findByIdReadModel(id: WalletId): WalletReadModel?

    suspend fun save(wallet: Wallet): Result<Wallet>
}
