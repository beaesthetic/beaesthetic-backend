package it.beaesthetic.wallet.infra

import io.quarkus.mongodb.panache.common.MongoEntity
import io.quarkus.mongodb.panache.kotlin.PanacheMongoEntityBase
import io.quarkus.runtime.annotations.RegisterForReflection
import java.time.Instant
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonId
import org.bson.codecs.pojo.annotations.BsonProperty

@RegisterForReflection
@MongoEntity(collection = "pending_gift_cards")
data class PendingGiftCardEntity
@BsonCreator
constructor(
    @BsonId val id: String,
    @BsonProperty("code") val code: String,
    @BsonProperty("amount") val amount: Double,
    @BsonProperty("createdAt") val createdAt: Instant,
    @BsonProperty("expiresAt") val expiresAt: Instant,
) : PanacheMongoEntityBase()
