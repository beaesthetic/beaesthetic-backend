package it.beaesthetic.appointment.agenda.infra.mongo

import com.mongodb.client.model.Filters
import com.mongodb.client.model.IndexModel
import com.mongodb.client.model.Indexes
import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.quarkus.runtime.Startup
import io.smallrye.mutiny.Uni
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.domain.*
import it.beaesthetic.appointment.common.OptimisticConcurrency
import it.beaesthetic.appointment.common.OptimisticLockException
import it.beaesthetic.appointment.common.panache.PanacheUtils.updateOne
import jakarta.enterprise.context.ApplicationScoped
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
class MongoAgendaRepository(private val panacheAgendaRepository: PanacheAgendaRepository) :
    AgendaRepository {
    override suspend fun findSchedule(
        scheduleId: String
    ): OptimisticConcurrency.VersionedEntity<AgendaSchedule>? {
        return panacheAgendaRepository
            .find("_id", scheduleId)
            .firstResult()
            .map {
                if (it != null)
                    OptimisticConcurrency.VersionedEntity(EntityMapper.toDomain(it), it.version)
                else null
            }
            .awaitSuspending()
    }

    override suspend fun saveSchedule(
        schedule: AgendaSchedule,
        expectedVersion: Long
    ): Result<AgendaSchedule> =
        kotlin.runCatching {
            val entity =
                EntityMapper.toEntity(schedule, expectedVersion).let {
                    it.copy(version = it.version + 1)
                }
            val filter =
                Filters.and(Filters.eq("_id", schedule.id), Filters.eq("version", expectedVersion))
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
                                Uni.createFrom().failure(OptimisticLockException(expectedVersion))
                            } else {
                                Uni.createFrom().item(entity)
                            }
                        }
                        .map { EntityMapper.toDomain(it) }
                        .awaitSuspending()
            }
        }

    override suspend fun findSchedules(timeSpan: TimeSpan): List<AgendaSchedule> {
        val schedules =
            panacheAgendaRepository.find(
                """
                {
                  ${'$'}or: [
                   { start: { ${'$'}gte: ?1 } },
                   { end: { ${'$'}lte: ?2 } }
                  ],
                  'cancelReason': { ${'$'}exists: false }
                 }
            """
                    .trimIndent(),
                timeSpan.start,
                timeSpan.end
            )
        return schedules
            .stream()
            .map { EntityMapper.toDomain(it) }
            .collect()
            .asList()
            .awaitSuspending()
    }

    override suspend fun findByAttendeeId(attendeeId: String): List<AgendaSchedule> {
        return panacheAgendaRepository
            .find("attendee._id == ?1", attendeeId)
            .stream()
            .map { EntityMapper.toDomain(it) }
            .collect()
            .asList()
            .awaitSuspending()
    }

    override suspend fun deleteSchedule(scheduleId: String): Result<Boolean> {
        TODO("Not yet implemented")
    }
}
