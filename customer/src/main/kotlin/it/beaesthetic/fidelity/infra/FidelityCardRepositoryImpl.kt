package it.beaesthetic.fidelity.infra

import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.fidelity.domain.*
import it.beaesthetic.fidelity.infra.FidelityCardRepositoryImpl.EntityMapper.toDomain
import jakarta.enterprise.context.ApplicationScoped
import org.bson.Document
import java.time.Instant

@ApplicationScoped
class PanacheFidelityCardRepository : ReactivePanacheMongoRepository<FidelityCardEntity>

@ApplicationScoped
class FidelityCardRepositoryImpl(
    private val panacheFidelityCardRepository: PanacheFidelityCardRepository
) : FidelityCardRepository {
    override suspend fun findAll(): List<FidelityCard> {
        return panacheFidelityCardRepository.findAll()
            .list().awaitSuspending().map { toDomain(it) }
    }

    override suspend fun findById(id: String): FidelityCard? {
        return panacheFidelityCardRepository.find("_id", id)
            .firstResult()
            .map { entity -> entity?.let { toDomain(entity) } }
            .awaitSuspending()
    }

    override suspend fun findByVoucherId(voucherId: String): FidelityCard? {
        return panacheFidelityCardRepository.find(
            Document("vouchers", Document("\$elemMatch", Document("id", voucherId)))
        )
            .firstResult()
            .map { entity -> entity?.let { toDomain(entity) } }
            .awaitSuspending()
    }

    override suspend fun findByCustomerId(customerId: CustomerId): FidelityCard? {
        return panacheFidelityCardRepository.find("customerId", customerId.id)
            .firstResult()
            .map { entity -> entity?.let { toDomain(entity) } }
            .awaitSuspending()
    }

    override suspend fun save(card: FidelityCard): Result<FidelityCard> {
        return Result.runCatching {
            panacheFidelityCardRepository.persistOrUpdate(
                FidelityCardEntity(
                    id = card.id,
                    customerId = card.customerId.id,
                    solariumPurchases = card.purchasesOf(FidelityTreatment.SOLARIUM),
                    vouchers = card.availableVouchers.map {
                        VoucherItem(
                            id = it.id.value,
                            treatment = it.treatment,
                            amount = null,
                            type = "free",
                            isUsed = it.isUsed,
                            createdAt = Instant.now()
                        )
                    },
                    createdAt = Instant.now(),
                    updatedAt = Instant.now(),
                )
            ).map { card }.awaitSuspending()
        }
    }

    private object EntityMapper {
        fun toDomain(entity: FidelityCardEntity) = FidelityCard(
            id = entity.id,
            customerId = CustomerId(entity.customerId),
            vouchers = entity.vouchers.map {
                Voucher(
                    id = VoucherId(it.id),
                    createdAt = it.createdAt,
                    isUsed = it.isUsed,
                    treatment = it.treatment,
                )
            },
            purchaseCounter = mapOf(
                FidelityTreatment.SOLARIUM.name to entity.solariumPurchases
            )
        )
    }
}