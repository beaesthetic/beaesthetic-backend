package it.beaesthetic.notification.domain

import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonProperty

data class ChannelMetadata
@BsonCreator
constructor(@BsonProperty("providerResourceId") val providerResourceId: String)
