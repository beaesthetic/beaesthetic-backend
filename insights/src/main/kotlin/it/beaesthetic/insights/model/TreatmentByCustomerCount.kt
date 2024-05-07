package it.beaesthetic.insights.model

import io.quarkus.mongodb.panache.common.MongoEntity

@MongoEntity(collection="TreatmentByCustomerCount")
data class TreatmentByCustomerCount(
    val serviceName: String,
    val attendeeId: String,
    val count: Int,
    val time: Long
)