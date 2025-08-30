package it.beaesthetic.wallet.infra

import com.mongodb.client.model.Aggregates
import com.mongodb.client.model.Filters
import com.mongodb.client.model.Projections
import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.wallet.application.read.WalletReadModel
import it.beaesthetic.wallet.domain.Wallet
import it.beaesthetic.wallet.domain.WalletId
import it.beaesthetic.wallet.domain.WalletRepository
import it.beaesthetic.wallet.infra.mappers.WalletEntityMapper
import jakarta.enterprise.context.ApplicationScoped

@ApplicationScoped class PanacheWalletRepository : ReactivePanacheMongoRepository<WalletEntity>

@ApplicationScoped
class WalletRepositoryImpl(
    private val mapper: WalletEntityMapper,
    private val panacheWalletRepository: PanacheWalletRepository,
) : WalletRepository {

    private val joinPipeline =
        listOf(
            Aggregates.lookup("customers", "owner", "_id", "customer"),
            Aggregates.unwind("\$customer"),
            Aggregates.project(
                Projections.fields(
                    Projections.excludeId(),
                    Projections.computed("wallet", "\$\$ROOT"),
                    Projections.computed("customer", "\$customer"),
                )
            ),
            Aggregates.unset("wallet.customer"),
        )

    override suspend fun findAll(): List<WalletReadModel> {
        return panacheWalletRepository
            .mongoCollection()
            .aggregate(joinPipeline, WalletReadModel::class.java)
            .collect()
            .asList()
            .awaitSuspending()
    }

    override suspend fun findByCustomer(owner: String): Wallet? {
        return panacheWalletRepository
            .find("owner", owner)
            .firstResult()
            .map { it?.let { mapper.entityToWallet(it) } }
            .awaitSuspending()
    }

    override suspend fun findById(id: WalletId): Wallet? {
        return panacheWalletRepository
            .find("_id", id.value)
            .firstResult()
            .map { it?.let { mapper.entityToWallet(it) } }
            .awaitSuspending()
    }

    override suspend fun findByIdReadModel(id: WalletId): WalletReadModel? {
        val pipeline = listOf(Aggregates.match(Filters.eq("_id", id.value))) + joinPipeline
        return panacheWalletRepository
            .mongoCollection()
            .aggregate(pipeline, WalletReadModel::class.java)
            .collect()
            .asList()
            .awaitSuspending()
            .firstOrNull()
    }

    override suspend fun save(wallet: Wallet): Result<Wallet> =
        Result.runCatching {
            panacheWalletRepository
                .persistOrUpdate(mapper.walletToEntity(wallet))
                .map { wallet }
                .awaitSuspending()
        }
}
