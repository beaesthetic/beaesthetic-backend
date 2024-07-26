package it.beaesthetic.wallet.infra

import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.smallrye.mutiny.coroutines.awaitSuspending
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

    override suspend fun findAll(): List<Wallet> {
        return panacheWalletRepository.findAll().list().awaitSuspending().map {
            mapper.entityToWallet(it)
        }
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

    override suspend fun save(wallet: Wallet): Result<Wallet> =
        Result.runCatching {
            panacheWalletRepository
                .persistOrUpdate(mapper.walletToEntity(wallet))
                .map { wallet }
                .awaitSuspending()
        }
}
