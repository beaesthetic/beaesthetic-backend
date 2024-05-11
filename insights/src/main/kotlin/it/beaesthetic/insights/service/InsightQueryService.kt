package it.beaesthetic.insights.service

import com.mongodb.client.model.Filters
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.insights.model.TreatmentByCustomerCount
import it.beaesthetic.insights.model.TreatmentUsageCount
import it.beaesthetic.insights.repository.TreatmentCustomerCountRepository
import it.beaesthetic.insights.repository.TreatmentUsageCountRepository
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.asFlow
import kotlinx.coroutines.flow.flowOf
import java.time.Instant

class InsightQueryService(
    private val treatmentUsageCountRepository: TreatmentUsageCountRepository,
    private val treatmentCustomerCountRepository: TreatmentCustomerCountRepository
) {

    suspend fun getAttendeesTreatmentsUsages(
        start: Instant,
        end: Instant,
        attendeeId: String?
    ): Flow<TreatmentByCustomerCount> = when(attendeeId) {
        null -> treatmentCustomerCountRepository.list(
            "time >= ?1 and time <= ?2", start, end
        )
        else -> treatmentCustomerCountRepository.list(
            "time >= ?1 and time <= ?2 and attendeeId = ?3",
            start, end, attendeeId
        )
    }.awaitSuspending().asFlow()

    suspend fun getTreatmentsUsages(
        start: Instant,
        end: Instant
    ): Flow<TreatmentUsageCount> {
        return treatmentUsageCountRepository.list("time >= ?1 and time <= ?2", start, end)
            .awaitSuspending()
            .asFlow()
    }
}