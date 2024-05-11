package it.beaesthetic.insights.adapter

import io.smallrye.mutiny.Uni
import io.smallrye.mutiny.coroutines.uni
import io.vertx.core.Vertx
import io.vertx.kotlin.coroutines.dispatcher
import it.beaesthetic.generated.insights.api.TreatmentsApi
import it.beaesthetic.generated.insights.api.model.TreatmentsByDayCustomerDto
import it.beaesthetic.generated.insights.api.model.TreatmentsByDayDto
import it.beaesthetic.insights.service.InsightQueryService
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.toList
import java.time.Duration
import java.time.OffsetDateTime
import java.time.ZoneOffset
import java.util.*

@OptIn(ExperimentalCoroutinesApi::class)
class InsightsRestResource(
    private val service: InsightQueryService
) : TreatmentsApi {

    private val coroutineScope: CoroutineScope = CoroutineScope(Vertx.currentContext().dispatcher())

    private val utcZone by lazy {
        ZoneOffset.UTC
    }

    override fun attendeesTreatmentsUsagesGet(
        startTime: OffsetDateTime?,
        endTime: OffsetDateTime?,
        attendeeId: UUID?
    ): Uni<List<TreatmentsByDayCustomerDto>> = uni(coroutineScope) {
        val end = endTime ?: OffsetDateTime.now()
        val start = startTime ?: (end - Duration.ofMinutes(60))

        service.getAttendeesTreatmentsUsages(
            start.toInstant(),
            end.toInstant(),
            attendeeId?.toString()
        ).map { TreatmentsByDayCustomerDto()
            .serviceName(it.serviceName)
            .count(it.count)
            .attendeeId(UUID.fromString(it.attendeeId))
            .time(it.time.atOffset(utcZone))
        }.toList()
    }

    override fun treatmentsUsagesGet(
        startTime: OffsetDateTime?,
        endTime: OffsetDateTime?
    ): Uni<List<TreatmentsByDayDto>> = uni(coroutineScope) {
        val end = endTime ?: OffsetDateTime.now()
        val start = startTime ?: (end - Duration.ofMinutes(60))
        service.getTreatmentsUsages(start.toInstant(), end.toInstant())
            .map {
                TreatmentsByDayDto()
                    .serviceName(it.serviceName)
                    .count(it.count)
                    .time(it.time.atOffset(utcZone))
            }
            .toList()
    }
}