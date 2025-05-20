package it.beaesthetic.customer.infra

import io.quarkus.mongodb.panache.common.MongoEntity
import io.quarkus.mongodb.panache.kotlin.PanacheMongoEntityBase
import io.quarkus.runtime.annotations.RegisterForReflection
import java.time.Instant
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonId
import org.bson.codecs.pojo.annotations.BsonProperty

@RegisterForReflection
@MongoEntity(collection = "customers")
data class CustomerEntity
@BsonCreator
constructor(
    @BsonId val id: String,
    @BsonProperty("name") val name: String,
    @BsonProperty("surname") val surname: String? = null,
    @BsonProperty("email") val email: String? = null,
    @BsonProperty("phone") val phone: String? = null,
    @BsonProperty("note") val note: String,
    @BsonProperty("searchGrams") val searchGrams: String? = null,
    @BsonProperty("updatedAt") val updatedAt: Instant,
) : PanacheMongoEntityBase()

@RegisterForReflection
@MongoEntity(collection = "delete_customers")
data class DeletedCustomerEntity
@BsonCreator
constructor(
    @BsonId val id: String,
    @BsonProperty("name") val name: String,
    @BsonProperty("surname") val surname: String? = null,
    @BsonProperty("email") val email: String? = null,
    @BsonProperty("phone") val phone: String? = null,
    @BsonProperty("note") val note: String,
    @BsonProperty("searchGrams") val searchGrams: String? = null,
    @BsonProperty("updatedAt") val updatedAt: Instant,
    @BsonProperty("deletedAt") val deletedAt: Instant? = null,
) : PanacheMongoEntityBase()
