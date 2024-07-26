package it.beaesthetic.wallet.rest.serialization

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonSubTypes
import com.fasterxml.jackson.annotation.JsonTypeInfo
import it.beaesthetic.wallet.generated.api.model.GiftCardMoneyCreditedEventDto
import it.beaesthetic.wallet.generated.api.model.GiftCardMoneyExpiredEventDto
import it.beaesthetic.wallet.generated.api.model.MoneyChargedEventDto
import it.beaesthetic.wallet.generated.api.model.MoneyCreditedEventDto

@JsonIgnoreProperties(value = ["type"], allowSetters = true)
@JsonTypeInfo(
    use = JsonTypeInfo.Id.NAME,
    property = "type",
    include = JsonTypeInfo.As.EXTERNAL_PROPERTY,
    visible = true
)
@JsonSubTypes(
    value =
        [
            JsonSubTypes.Type(
                value = GiftCardMoneyExpiredEventDto::class,
                name = "GiftCardMoneyExpired"
            ),
            JsonSubTypes.Type(
                value = GiftCardMoneyCreditedEventDto::class,
                name = "GiftCardMoneyCredited"
            ),
            JsonSubTypes.Type(value = MoneyCreditedEventDto::class, name = "MoneyCredited"),
            JsonSubTypes.Type(value = MoneyChargedEventDto::class, name = "MoneyCharged"),
        ]
)
interface WalletEventDtoMixin
