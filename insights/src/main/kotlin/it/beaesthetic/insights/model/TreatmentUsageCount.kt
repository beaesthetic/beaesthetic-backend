package it.beaesthetic.insights.model

import io.quarkus.mongodb.panache.common.MongoEntity
import io.quarkus.mongodb.panache.kotlin.PanacheMongoEntity
import io.quarkus.runtime.annotations.RegisterForReflection
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonProperty
import java.time.Instant

@RegisterForReflection
@MongoEntity(collection="treatmentUsageCount")
data class TreatmentUsageCount @BsonCreator constructor(
    @BsonProperty("serviceName") val serviceName: String,
    @BsonProperty("count") val count: Int,
    @BsonProperty("time") val time: Instant,
    @BsonProperty("updatedAt") val updatedAt: Instant
) : PanacheMongoEntity()