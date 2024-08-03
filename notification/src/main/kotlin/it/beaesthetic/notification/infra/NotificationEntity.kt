package it.beaesthetic.notification.infra

import io.quarkus.mongodb.panache.common.MongoEntity
import io.quarkus.mongodb.panache.kotlin.PanacheMongoEntityBase
import io.quarkus.runtime.annotations.RegisterForReflection
import it.beaesthetic.notification.domain.Channel
import it.beaesthetic.notification.domain.ChannelMetadata
import java.time.Instant
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonId
import org.bson.codecs.pojo.annotations.BsonProperty

@RegisterForReflection(registerFullHierarchy = true)
@MongoEntity(collection = "notifications")
data class NotificationEntity
@BsonCreator
constructor(
    @BsonId val id: String,
    @BsonProperty("title") val title: String,
    @BsonProperty("content") val content: String,
    @get:BsonProperty("isSent") @param:BsonProperty("isSent") val isSent: Boolean,
    @get:BsonProperty("isSentConfirmed")
    @param:BsonProperty("isSentConfirmed")
    val isSentConfirmed: Boolean,
    @BsonProperty("channel") val channel: Channel,
    @BsonProperty("channelData") val channelData: ChannelMetadata? = null,
    @BsonProperty("createdAt") val createdAt: Instant,
    @BsonProperty("updatedAt") val updatedAt: Instant
) : PanacheMongoEntityBase()
