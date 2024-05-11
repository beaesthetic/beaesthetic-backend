package it.beaesthetic.insights.service

import io.quarkus.mongodb.reactive.ReactiveMongoClient
import io.quarkus.panache.common.Sort
import io.smallrye.mutiny.coroutines.asFlow
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.insights.model.TreatmentByCustomerCount
import it.beaesthetic.insights.model.TreatmentUsageCount
import it.beaesthetic.insights.repository.Aggregates
import it.beaesthetic.insights.repository.TreatmentCustomerCountRepository
import it.beaesthetic.insights.repository.TreatmentUsageCountRepository
import kotlinx.coroutines.flow.map
import org.jboss.logging.Logger
import java.time.Duration
import java.time.Instant

class InsightService(
    private val appointmentMongoClient: ReactiveMongoClient,
    private val treatmentUsageCountRepository: TreatmentUsageCountRepository,
    private val treatmentCustomerCountRepository: TreatmentCustomerCountRepository
) {

    private val logger = Logger.getLogger(InsightService::class.java)

    companion object {
        const val DATABASE = "appointment"
        const val ACTIVITY_COLLECTION = "agendaactivities"
    }

    suspend fun computeMostUsedTreatments(until: Instant) {
        val startFrom = getMostUsedTreatmentsLastProcessTime() + Duration.ofDays(1)
        val query = Aggregates.templatizeQuery(
            "/aggregation/mostUsageTreatments.json.tpl",
            "startDate" to startFrom,
            "endDate" to until
        )
        logger.infof("Computing most used treatments startFrom: [%s], endTo: [%s]", startFrom, until)
        appointmentMongoClient.getDatabase(DATABASE)
            .getCollection(ACTIVITY_COLLECTION)
            .aggregate(query)
            .asFlow()
            .map {
                TreatmentUsageCount(
                    serviceName = it.getString("serviceName"),
                    time = it.getDate("date").toInstant(),
                    count = it.getInteger("count"),
                    updatedAt = Instant.now()
                )
            }
            .collect {
                treatmentUsageCountRepository.persist(it).awaitSuspending()
            }
    }

    suspend fun computeTreatmentsByCustomer(until: Instant) {
        val startFrom = getTreatmentsByCustomerLastProcessTime() + Duration.ofDays(1)
        val query = Aggregates.templatizeQuery(
            "/aggregation/treatmentsByCustomer.json.tpl",
            "startDate" to startFrom,
            "endDate" to until
        )
        logger.infof("Computing treatments by customer startFrom: [%s], endTo: [%s]", startFrom, until)
        appointmentMongoClient.getDatabase(DATABASE)
            .getCollection(ACTIVITY_COLLECTION)
            .aggregate(query)
            .asFlow()
            .map {
                TreatmentByCustomerCount(
                    serviceName = it.getString("serviceName"),
                    attendeeId = it.getString("attendeeId"),
                    time = it.getDate("date").toInstant(),
                    count = it.getInteger("count"),
                    updatedAt = Instant.now()
                )
            }
            .collect {
                treatmentCustomerCountRepository.persist(it).awaitSuspending()
            }
    }

    private suspend fun getMostUsedTreatmentsLastProcessTime(): Instant {
        return treatmentUsageCountRepository
            .streamAll(Sort.by("updatedAt").descending())
            .toUni()
            .map { it?.updatedAt }
            .awaitSuspending() ?: Instant.EPOCH
    }

    private suspend fun getTreatmentsByCustomerLastProcessTime(): Instant {
        return treatmentCustomerCountRepository
            .streamAll(Sort.by("updatedAt").descending())
            .toUni()
            .map { it?.updatedAt }
            .awaitSuspending() ?: Instant.EPOCH
    }
}