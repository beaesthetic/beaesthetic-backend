package it.beaesthetic.wallet

import com.mongodb.client.model.IndexModel
import com.mongodb.client.model.IndexOptions
import com.mongodb.client.model.Indexes
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.common.MongoInitializer
import it.beaesthetic.wallet.application.WalletService
import it.beaesthetic.wallet.domain.WalletRepository
import it.beaesthetic.wallet.infra.PanacheWalletRepository
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces

@Dependent
class DependencyConfiguration {

    @Produces
    fun walletService(walletRepository: WalletRepository): WalletService {
        return WalletService(walletRepository)
    }

    @Produces
    fun mongoInitializer(panacheWalletRepository: PanacheWalletRepository): MongoInitializer =
        object : MongoInitializer {
            override suspend fun initialize() {
                panacheWalletRepository
                    .mongoCollection()
                    .createIndexes(
                        listOf(IndexModel(Indexes.ascending("owner"), IndexOptions().unique(true)))
                    )
                    .awaitSuspending()
            }
        }
}
