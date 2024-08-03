package it.beaesthetic.notification.driver.rest.dto

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonSubTypes
import com.fasterxml.jackson.annotation.JsonTypeInfo
import it.beaesthetic.notification.generated.api.model.EmailChannelDto
import it.beaesthetic.notification.generated.api.model.SmsChannelDto
import it.beaesthetic.notification.generated.api.model.WhatsappChannelDto

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
            JsonSubTypes.Type(value = EmailChannelDto::class, name = "email"),
            JsonSubTypes.Type(value = SmsChannelDto::class, name = "sms"),
            JsonSubTypes.Type(value = WhatsappChannelDto::class, name = "whatsapp"),
        ]
)
interface ChannelMixin
