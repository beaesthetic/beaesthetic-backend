package it.beaesthetic.insights.model

import io.quarkus.mongodb.panache.common.MongoEntity

@MongoEntity(collection="TreatmentUsageCount")
data class TreatmentUsageCount(
    val serviceName: String,
    val count: Int,
    val time: Long
)