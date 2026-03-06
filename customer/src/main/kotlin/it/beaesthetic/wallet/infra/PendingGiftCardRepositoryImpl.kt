package it.beaesthetic.wallet.infra

import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.wallet.domain.Money
import it.beaesthetic.wallet.domain.PendingGiftCard
import it.beaesthetic.wallet.domain.PendingGiftCardRepository
import jakarta.enterprise.context.ApplicationScoped

@ApplicationScoped
class PanachePendingGiftCardRepository :
    ReactivePanacheMongoRepository<PendingGiftCardEntity>

@ApplicationScoped
class PendingGiftCardRepositoryImpl(
    private val panache: PanachePendingGiftCardRepository,
) : PendingGiftCardRepository {

    override suspend fun save(card: PendingGiftCard): Result<PendingGiftCard> =
        Result.runCatching {
            panache.persistOrUpdate(card.toEntity()).map { card }.awaitSuspending()
        }

    override suspend fun findById(id: String): PendingGiftCard? =
        panache.find("_id", id).firstResult().map { it?.toDomain() }.awaitSuspending()

    override suspend fun findByCode(code: String): PendingGiftCard? =
        panache.find("code", code).firstResult().map { it?.toDomain() }.awaitSuspending()

    override suspend fun findAll(): List<PendingGiftCard> =
        panache.findAll().list().map { list -> list.map { it.toDomain() } }.awaitSuspending()

    override suspend fun findPage(limit: Int, offset: Int): List<PendingGiftCard> =
        panache.findAll().page(offset / limit, limit).list()
            .map { list -> list.map { it.toDomain() } }
            .awaitSuspending()

    override suspend fun delete(id: String) {
        panache.delete("_id", id).awaitSuspending()
    }
}

private fun PendingGiftCard.toEntity() =
    PendingGiftCardEntity(id, code, amount.amount, createdAt, expiresAt)

private fun PendingGiftCardEntity.toDomain() =
    PendingGiftCard(id, code, Money(amount), createdAt, expiresAt)
