package it.beaesthetic.wallet.infra

import io.quarkus.mongodb.panache.common.MongoEntity
import io.quarkus.mongodb.panache.kotlin.PanacheMongoEntityBase
import io.quarkus.runtime.annotations.RegisterForReflection
import java.time.Instant
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonDiscriminator
import org.bson.codecs.pojo.annotations.BsonProperty

@RegisterForReflection
@MongoEntity(collection = "wallets")
data class WalletEntity
@BsonCreator
constructor(
    @BsonProperty("id") val id: String, // INDEX
    @BsonProperty("owner") val owner: String, // INDEX
    @BsonProperty("availableAmount") val availableAmount: Double,
    @BsonProperty("spentAmount") val spentAmount: Double,
    @BsonProperty("operations") val operations: List<WalletEventEntity>,
    @BsonProperty("activeGiftCards") val activeGiftCards: List<GiftCardEntity>,
    @BsonProperty("createdAt") val createdAt: Instant,
    @BsonProperty("updatedAt") val updatedAt: Instant,
) : PanacheMongoEntityBase()

@RegisterForReflection
data class GiftCardEntity
@BsonCreator
constructor(
    @BsonProperty("id") val id: String, // INDEX
    @BsonProperty("customerId") val customerId: String,
    @BsonProperty("amount") val amount: Double,
    @BsonProperty("availableAmount") val availableAmount: Double,
    @BsonProperty("amountSpent") val amountSpent: Double,
    @BsonProperty("createdAt") val createdAt: Instant,
    @BsonProperty("expireAt") val expireAt: Instant,
)

@RegisterForReflection @BsonDiscriminator(key = "type") sealed interface WalletEventEntity

@RegisterForReflection
@BsonDiscriminator(key = "type", value = "MoneyCredited")
data class MoneyCreditedEntity
@BsonCreator
constructor(
    @BsonProperty("at") val at: Instant,
    @BsonProperty("amount") val amount: Double,
) : WalletEventEntity

@RegisterForReflection
@BsonDiscriminator(key = "type", value = "GiftCardMoneyCredited")
data class GiftCardMoneyCreditedEntity
@BsonCreator
constructor(
    @BsonProperty("at") val at: Instant,
    @BsonProperty("giftCardId") val giftCardId: String,
    @BsonProperty("amount") val amount: Double,
    @get:BsonProperty("expireAt") @param:BsonProperty("expireAt") val expiresAt: Instant,
) : WalletEventEntity

@RegisterForReflection
@BsonDiscriminator(key = "type", value = "GiftCardMoneyExpired")
data class GiftCardMoneyExpiredEntity
@BsonCreator
constructor(
    @BsonProperty("at") val at: Instant,
    @BsonProperty("giftCardId") val giftCardId: String,
) : WalletEventEntity

@RegisterForReflection
@BsonDiscriminator(key = "type", value = "MoneyCharged")
data class MoneyChargedEntity
@BsonCreator
constructor(
    @BsonProperty("at") val at: Instant,
    @BsonProperty("amount") val amount: Double,
) : WalletEventEntity
