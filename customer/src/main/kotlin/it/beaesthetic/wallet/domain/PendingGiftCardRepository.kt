package it.beaesthetic.wallet.domain

interface PendingGiftCardRepository {
    suspend fun save(card: PendingGiftCard): Result<PendingGiftCard>
    suspend fun findById(id: String): PendingGiftCard?
    suspend fun findByCode(code: String): PendingGiftCard?
    suspend fun findAll(): List<PendingGiftCard>
    suspend fun findPage(limit: Int, offset: Int): List<PendingGiftCard>
    suspend fun delete(id: String)
}
