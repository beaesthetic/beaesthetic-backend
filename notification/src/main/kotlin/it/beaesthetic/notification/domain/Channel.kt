package it.beaesthetic.notification.domain

import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonProperty

sealed interface Channel

data class Email @BsonCreator constructor(
    @BsonProperty("email") val email: String
) : Channel

data class Sms @BsonCreator constructor(
    @BsonProperty("phone") val phone: String
) : Channel

data class WhatsApp @BsonCreator constructor(
    @BsonProperty("phone") val phone: String
) : Channel
