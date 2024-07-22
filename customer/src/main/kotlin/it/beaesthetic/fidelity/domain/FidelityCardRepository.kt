package it.beaesthetic.fidelity.domain

interface FidelityCardRepository {
    suspend fun findAll(): List<FidelityCard>
    suspend fun findById(id: String): FidelityCard?
    suspend fun findByVoucherId(voucherId: String): FidelityCard?
    suspend fun findByCustomerId(customerId: CustomerId): FidelityCard?
    suspend fun save(card: FidelityCard): Result<FidelityCard>
}