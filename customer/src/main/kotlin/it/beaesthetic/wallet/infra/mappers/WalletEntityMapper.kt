package it.beaesthetic.wallet.infra.mappers

import it.beaesthetic.wallet.domain.Money
import it.beaesthetic.wallet.domain.Wallet
import it.beaesthetic.wallet.infra.WalletEntity
import org.mapstruct.InheritInverseConfiguration
import org.mapstruct.Mapper
import org.mapstruct.Mapping

@Mapper(uses = [WalletEventEntityMapper::class, MoneyMapper::class, GiftCardMapper::class])
interface WalletEntityMapper {

    @Mapping(source = "giftCards", target = "activeGiftCards")
    @Mapping(target = "updatedAt", expression = "java(Instant.now())")
    fun walletToEntity(wallet: Wallet): WalletEntity

    @InheritInverseConfiguration
    @Mapping(target = "processorPolicy", expression = "java(new GiftCardProcessorPolicy())")
    fun entityToWallet(entity: WalletEntity): Wallet
}

@Mapper
abstract class MoneyMapper {
    fun unwrapMoney(money: Money): Double = money.amount
    fun wrapMoney(money: Double): Money = Money(money)
}
