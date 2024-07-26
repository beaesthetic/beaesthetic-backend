package it.beaesthetic.wallet.infra.mappers

import it.beaesthetic.wallet.domain.GiftCard
import it.beaesthetic.wallet.infra.GiftCardEntity
import org.mapstruct.InheritInverseConfiguration
import org.mapstruct.Mapper
import org.mapstruct.Mapping

@Mapper(componentModel = "cdi", uses = [MoneyMapper::class])
interface GiftCardMapper {
    @Mapping(source = "owner", target = "customerId")
    @Mapping(source = "expiresAt", target = "expireAt")
    fun giftCardToEntity(giftCard: GiftCard): GiftCardEntity

    @InheritInverseConfiguration fun entityToGiftCard(entity: GiftCardEntity): GiftCard
}
