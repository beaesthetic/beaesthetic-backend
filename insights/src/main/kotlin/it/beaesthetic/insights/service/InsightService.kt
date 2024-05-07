package it.beaesthetic.insights.service

import io.quarkus.mongodb.reactive.ReactiveMongoClient
import io.smallrye.mutiny.coroutines.asFlow
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.insights.model.TreatmentByCustomerCount
import it.beaesthetic.insights.model.TreatmentUsageCount
import it.beaesthetic.insights.repository.TreatmentCustomerCountRepository
import it.beaesthetic.insights.repository.TreatmentUsageCountRepository
import kotlinx.coroutines.flow.map
import org.bson.BsonArray

class InsightService(
    private val appointmentMongoClient: ReactiveMongoClient,
    private val treatmentUsageCountRepository: TreatmentUsageCountRepository,
    private val treatmentCustomerCountRepository: TreatmentCustomerCountRepository
) {

    companion object {
        const val DATABASE = "appointment"
        const val ACTIVITY_COLLECTION = "agendaactivities"
    }

    private val mostUsageTreatmentsQuery by lazy {
        InsightService::class.java.getResourceAsStream(
            "/aggregation/mostUsageTreatments.json"
        )?.bufferedReader()?.readText()?.let {
            BsonArray.parse(it)
        }?.map { it.asDocument() } ?: emptyList()
    }

    private val treatmentsByCustomerQuery by lazy {
        InsightService::class.java.getResourceAsStream(
            "/aggregation/treatmentsByCustomer.json"
        )?.bufferedReader()?.readText()?.let {
            BsonArray.parse(it)
        }?.map { it.asDocument() } ?: emptyList()
    }

    suspend fun computeMostUsedTreatments() {
        appointmentMongoClient.getDatabase(DATABASE)
            .getCollection(ACTIVITY_COLLECTION)
            .aggregate(mostUsageTreatmentsQuery)
            .asFlow()
            .map {
                TreatmentUsageCount(
                    serviceName = it.getString("serviceName"),
                    time = it.getDate("date").toInstant().epochSecond,
                    count = it.getInteger("count")
                )
            }
            .collect {
                treatmentUsageCountRepository.persist(it).awaitSuspending()
            }
    }

    suspend fun computeTreatmentsByCustomer() {
        appointmentMongoClient.getDatabase(DATABASE)
            .getCollection(ACTIVITY_COLLECTION)
            .aggregate(treatmentsByCustomerQuery)
            .asFlow()
            .map {
                TreatmentByCustomerCount(
                    serviceName = it.getString("serviceName"),
                    attendeeId = it.getString("attendeeId"),
                    time = it.getDate("date").toInstant().epochSecond,
                    count = it.getInteger("count")
                )
            }
            .collect {
                treatmentCustomerCountRepository.persist(it).awaitSuspending()
            }
    }

}