package it.beaesthetic.appointment.agenda.infra.mongo

import com.mongodb.client.model.Filters
import com.mongodb.client.model.IndexModel
import com.mongodb.client.model.Indexes
import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.quarkus.runtime.Startup
import io.smallrye.mutiny.Uni
import io.smallrye.mutiny.coroutines.awaitSuspending
import io.vertx.mutiny.core.eventbus.EventBus
import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.event.TimeSpan
import it.beaesthetic.appointment.common.OptimisticConcurrency
import it.beaesthetic.appointment.common.OptimisticLockException
import it.beaesthetic.appointment.common.panache.PanacheUtils.updateOne
import jakarta.enterprise.context.ApplicationScoped
import javax.swing.text.Document
import kotlinx.coroutines.runBlocking

@ApplicationScoped
class PanacheAgendaRepository : ReactivePanacheMongoRepository<AgendaEntity> {
    @Startup
    fun initializeIndexes() = runBlocking {
        mongoCollection()
            .createIndexes(
                listOf(
                    IndexModel(Indexes.ascending("start")),
                    IndexModel(Indexes.ascending("end")),
                    IndexModel(Indexes.ascending("attendee.id")),
                    IndexModel(Indexes.ascending("cancelReason")),
                )
            )
            .awaitSuspending()
    }
}

@ApplicationScoped
class MongoAgendaRepository(
    private val panacheAgendaRepository: PanacheAgendaRepository,
    private val eventBus: EventBus,
) : AgendaRepository {
    override suspend fun findEvent(
        scheduleId: AgendaEventId
    ): OptimisticConcurrency.VersionedEntity<AgendaEvent>? {
        return panacheAgendaRepository
            .find("_id", scheduleId.value)
            .firstResult()
            .map {
                if (it != null)
                    OptimisticConcurrency.VersionedEntity(EntityMapper.toDomain(it), it.version)
                else null
            }
            .awaitSuspending()
    }

    override suspend fun saveEvent(
        schedule: AgendaEvent,
        expectedVersion: Long
    ): Result<AgendaEvent> =
        kotlin
            .runCatching {
                val entity =
                    EntityMapper.toEntity(schedule, expectedVersion).let {
                        it.copy(version = it.version + 1)
                    }
                val filter =
                    Filters.and(
                        Filters.eq("_id", schedule.id.value),
                        Filters.eq("version", expectedVersion)
                    )
                when (expectedVersion) {
                    0.toLong() ->
                        panacheAgendaRepository
                            .persist(entity)
                            .map { EntityMapper.toDomain(it) }
                            .awaitSuspending()
                    else ->
                        panacheAgendaRepository
                            .updateOne(filter, entity)
                            .flatMap {
                                if (it.matchedCount.toInt() == 0) {
                                    Uni.createFrom()
                                        .failure(OptimisticLockException(expectedVersion))
                                } else {
                                    Uni.createFrom().item(entity)
                                }
                            }
                            .map { EntityMapper.toDomain(it) }
                            .awaitSuspending()
                }
            }
            .also {
                schedule.events.forEach { eventBus.publish(it.javaClass.simpleName, it) }
                schedule.clearEvents()
            }

    override suspend fun findEvents(timeSpan: TimeSpan): List<AgendaEvent> {
        val query =
            org.bson
                .Document(
                    "start",
                    org.bson.Document("\$gte", timeSpan.start).append("\$lte", timeSpan.end)
                )
                .append("cancelReason", org.bson.Document("\$exists", "false"))
        val schedules = panacheAgendaRepository.find(query)

        return schedules
            .stream()
            .map { EntityMapper.toDomain(it) }
            .collect()
            .asList()
            .awaitSuspending()
    }

    override suspend fun findByAttendeeId(attendeeId: String): List<AgendaEvent> {
        return panacheAgendaRepository
            .find("attendee._id == ?1", attendeeId)
            .stream()
            .map { EntityMapper.toDomain(it) }
            .collect()
            .asList()
            .awaitSuspending()
    }

    override suspend fun deleteEvent(scheduleId: AgendaEventId): Result<Boolean> {
        TODO("Not yet implemented")
    }
}
