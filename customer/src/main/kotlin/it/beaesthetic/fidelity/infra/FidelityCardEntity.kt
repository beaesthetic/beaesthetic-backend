package it.beaesthetic.fidelity.infra

import io.quarkus.mongodb.panache.common.MongoEntity
import io.quarkus.mongodb.panache.kotlin.PanacheMongoEntityBase
import io.quarkus.runtime.annotations.RegisterForReflection
import it.beaesthetic.fidelity.domain.FidelityTreatment
import java.time.Instant
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonId
import org.bson.codecs.pojo.annotations.BsonProperty

@RegisterForReflection
@MongoEntity(collection = "fidelitycards")
data class FidelityCardEntity
@BsonCreator
constructor(
    @BsonId val id: String,
    @BsonProperty("customerId") val customerId: String,
    @BsonProperty("solariumPurchases") val solariumPurchases: Int,
    @BsonProperty("vouchers") val vouchers: List<VoucherItem>,
    @BsonProperty("createdAt") val createdAt: Instant,
    @BsonProperty("updatedAt") val updatedAt: Instant
) : PanacheMongoEntityBase()

@RegisterForReflection
data class VoucherItem
@BsonCreator
constructor(
    @BsonId val id: String,
    @get:BsonProperty("amount") @param:BsonProperty("amount") val amount: Int? = null,
    @BsonProperty("treatment") val treatment: FidelityTreatment,
    @get:BsonProperty("_type") @param:BsonProperty("_type") val type: String,
    @get:BsonProperty("isUsed") @param:BsonProperty("isUsed") val isUsed: Boolean,
    @BsonProperty("createdAt") val createdAt: Instant
)
