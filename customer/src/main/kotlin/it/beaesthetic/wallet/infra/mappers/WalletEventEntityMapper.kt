package it.beaesthetic.wallet.infra.mappers

import it.beaesthetic.wallet.domain.*
import it.beaesthetic.wallet.infra.*
import org.mapstruct.InheritInverseConfiguration
import org.mapstruct.Mapper
import org.mapstruct.SubclassExhaustiveStrategy
import org.mapstruct.SubclassMapping

@Mapper(
    componentModel = "cdi",
    uses = [MoneyMapper::class],
    subclassExhaustiveStrategy = SubclassExhaustiveStrategy.RUNTIME_EXCEPTION
)
interface WalletEventEntityMapper {

    @SubclassMapping(source = MoneyCredited::class, target = MoneyCreditedEntity::class)
    @SubclassMapping(
        source = GiftCardMoneyCredited::class,
        target = GiftCardMoneyCreditedEntity::class
    )
    @SubclassMapping(
        source = GiftCardMoneyExpired::class,
        target = GiftCardMoneyExpiredEntity::class
    )
    @SubclassMapping(source = MoneyCharge::class, target = MoneyChargedEntity::class)
    fun walletEventToEntity(walletEvent: WalletEvent): WalletEventEntity

    @InheritInverseConfiguration fun entityToWalletEvent(entity: WalletEventEntity): WalletEvent
}
