package it.beaesthetic.wallet.infra

import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.wallet.domain.GiftCardVoucher
import it.beaesthetic.wallet.domain.GiftCardVoucherRepository
import it.beaesthetic.wallet.domain.Money
import jakarta.enterprise.context.ApplicationScoped

@ApplicationScoped
class PanacheGiftCardVoucherRepository :
    ReactivePanacheMongoRepository<GiftCardVoucherEntity>

@ApplicationScoped
class GiftCardVoucherRepositoryImpl(
    private val panache: PanacheGiftCardVoucherRepository,
) : GiftCardVoucherRepository {

    override suspend fun save(voucher: GiftCardVoucher): Result<GiftCardVoucher> =
        Result.runCatching {
            panache.persistOrUpdate(voucher.toEntity()).map { voucher }.awaitSuspending()
        }

    override suspend fun findById(id: String): GiftCardVoucher? =
        panache.find("_id", id).firstResult().map { it?.toDomain() }.awaitSuspending()

    override suspend fun findByCode(code: String): GiftCardVoucher? =
        panache.find("code", code).firstResult().map { it?.toDomain() }.awaitSuspending()

    override suspend fun findAll(): List<GiftCardVoucher> =
        panache.findAll().list().map { list -> list.map { it.toDomain() } }.awaitSuspending()

    override suspend fun findPage(limit: Int, offset: Int): List<GiftCardVoucher> =
        panache.findAll().page(offset / limit, limit).list()
            .map { list -> list.map { it.toDomain() } }
            .awaitSuspending()

    override suspend fun delete(id: String) {
        panache.delete("_id", id).awaitSuspending()
    }
}

private fun GiftCardVoucher.toEntity() =
    GiftCardVoucherEntity(id, code, amount.amount, createdAt, expiresAt)

private fun GiftCardVoucherEntity.toDomain() =
    GiftCardVoucher(id, code, Money(amount), createdAt, expiresAt)
