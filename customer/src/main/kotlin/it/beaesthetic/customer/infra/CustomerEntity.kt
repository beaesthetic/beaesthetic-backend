package it.beaesthetic.customer.infra

import io.quarkus.mongodb.panache.common.MongoEntity
import io.quarkus.mongodb.panache.kotlin.PanacheMongoEntityBase
import io.quarkus.runtime.annotations.RegisterForReflection
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonId
import org.bson.codecs.pojo.annotations.BsonProperty
import java.time.Instant

@RegisterForReflection
@MongoEntity(collection = "customers")
data class CustomerEntity @BsonCreator constructor(
    @BsonId val id: String,
    @BsonProperty("name") val name: String,
    @BsonProperty("surname") val surname: String,
    @BsonProperty("email") val email: String? = null,
    @BsonProperty("phone") val phone: String? = null,
    @BsonProperty("note") val note: String,
    @BsonProperty("searchGrams") val searchGrams: String? = null,
    @BsonProperty("updatedAt") val updatedAt: Instant
) : PanacheMongoEntityBase()
