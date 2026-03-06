package it.beaesthetic.wallet

import com.mongodb.client.model.IndexModel
import com.mongodb.client.model.IndexOptions
import com.mongodb.client.model.Indexes
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.common.MongoInitializer
import it.beaesthetic.wallet.application.WalletService
import it.beaesthetic.wallet.domain.GiftCardVoucherRepository
import it.beaesthetic.wallet.domain.WalletRepository
import it.beaesthetic.wallet.http.GiftCardV2Resource
import it.beaesthetic.wallet.infra.PanacheGiftCardVoucherRepository
import it.beaesthetic.wallet.infra.PanacheWalletRepository
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces

@Dependent
class DependencyConfiguration {

    @Produces
    fun walletService(
        walletRepository: WalletRepository,
        voucherRepository: GiftCardVoucherRepository,
    ): WalletService = WalletService(walletRepository, voucherRepository)

    @Produces
    fun giftCardV2Resource(walletService: WalletService): GiftCardV2Resource =
        GiftCardV2Resource(walletService)

    @Produces
    fun mongoInitializer(
        panacheWalletRepository: PanacheWalletRepository,
        panacheGiftCardVoucherRepository: PanacheGiftCardVoucherRepository,
    ): MongoInitializer =
        object : MongoInitializer {
            override suspend fun initialize() {
                panacheWalletRepository
                    .mongoCollection()
                    .createIndexes(
                        listOf(IndexModel(Indexes.ascending("owner"), IndexOptions().unique(true)))
                    )
                    .awaitSuspending()
                panacheGiftCardVoucherRepository
                    .mongoCollection()
                    .createIndexes(
                        listOf(IndexModel(Indexes.ascending("code"), IndexOptions().unique(true)))
                    )
                    .awaitSuspending()
            }
        }
}
