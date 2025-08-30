package it.beaesthetic.fidelity.application.data

import io.quarkus.runtime.annotations.RegisterForReflection
import it.beaesthetic.fidelity.infra.VoucherItem
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonId
import org.bson.codecs.pojo.annotations.BsonProperty

@RegisterForReflection
data class FidelityCardReadModel
@BsonCreator
constructor(
    @BsonId val id: String,
    @BsonProperty("customer") val customer: CustomerInfo,
    @BsonProperty("vouchers") val vouchers: List<VoucherItem>,
    @BsonProperty("solariumPurchases") val solariumPurchases: Int,
)

@RegisterForReflection
data class CustomerInfo
@BsonCreator
constructor(
    @BsonId val id: String,
    @BsonProperty("name") val name: String?,
    @BsonProperty("surname") val surname: String?,
    @BsonProperty("email") val email: String?,
    @BsonProperty("phone") val phone: String?,
)
