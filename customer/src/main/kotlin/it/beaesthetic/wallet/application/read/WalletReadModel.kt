package it.beaesthetic.wallet.application.read

import io.quarkus.runtime.annotations.RegisterForReflection
import it.beaesthetic.wallet.infra.WalletEntity
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonId
import org.bson.codecs.pojo.annotations.BsonProperty

@RegisterForReflection
data class WalletReadModel
@BsonCreator
constructor(
    @BsonProperty("customer") val customer: CustomerInfo,
    @BsonProperty("wallet") val wallet: WalletEntity,
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
