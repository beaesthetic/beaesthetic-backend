package it.beaesthetic.appointment.agenda.infra

import com.mongodb.client.model.IndexModel
import com.mongodb.client.model.Indexes
import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.quarkus.panache.common.Parameters
import io.quarkus.runtime.Startup
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.domain.*
import it.beaesthetic.appointment.agenda.domain.AgendaScheduleData
import it.beaesthetic.appointment.common.OptimisticConcurrency
import jakarta.enterprise.context.ApplicationScoped
import kotlinx.coroutines.runBlocking
import org.bson.Document
import java.time.Instant

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
        private val panacheAgendaRepository: PanacheAgendaRepository
) : AgendaRepository {
    override suspend fun findSchedule(scheduleId: String): OptimisticConcurrency.VersionedEntity<AgendaSchedule>? {
        return panacheAgendaRepository.find("_id", scheduleId)
                .firstResult()
                .map {
                    if (it != null)
                        OptimisticConcurrency.VersionedEntity(EntityMapper.toDomain(it), it.version)
                    else null }
                .awaitSuspending()
    }

    override suspend fun saveSchedule(schedule: AgendaSchedule, expectedVersion: Long): Result<AgendaSchedule> {
        val entity = EntityMapper.toEntity(schedule, expectedVersion)
        /*panacheAgendaRepository.mongoCollection().u
                .updateOne(
                        Document("id", schedule.id),
                )*/
        TODO()
    }

    override suspend fun findSchedules(timeSpan: TimeSpan): List<AgendaSchedule> {
        val schedules = panacheAgendaRepository.find(
                "(start >= :start and start <= :end) or (end >= :end and end <= :end) and cancelReason == null",
                Parameters.with("start", timeSpan.start),
                Parameters.with("end", timeSpan.end),
        )
        return schedules.stream()
                .map { EntityMapper.toDomain(it) }
                .collect()
                .asList()
                .awaitSuspending()
    }

    override suspend fun findByAttendeeId(attendeeId: String): List<AgendaSchedule> {
        return panacheAgendaRepository.find("cancelReason == null")
                .stream()
                .map { EntityMapper.toDomain(it) }
                .collect()
                .asList()
                .awaitSuspending()
    }

    override suspend fun deleteSchedule(scheduleId: String): Result<Boolean> {
        TODO("Not yet implemented")
    }

    object EntityMapper {
        fun toEntity(scheduleAgenda: AgendaSchedule, version: Long): AgendaEntity {
            return AgendaEntity(
                    id = scheduleAgenda.id,
                    attendee = scheduleAgenda.attendee,
                    start = scheduleAgenda.timeSpan.start,
                    end = scheduleAgenda.timeSpan.end,
                    cancelReason = scheduleAgenda.cancelReason,
                    createdAt = scheduleAgenda.createdAt,
                    updatedAt = Instant.now(),
                    data = when(scheduleAgenda.data) {
                        is AppointmentScheduleData -> AgendaAppointmentData(
                                services = scheduleAgenda.data.services.map { it.name }
                        )
                        is BasicScheduleData -> AgendaBasicData(
                                title = scheduleAgenda.data.title,
                                description = scheduleAgenda.data.description ?: ""
                        )
                    },
                    version = version
            )
        }

        fun toDomain(scheduleAgendaEntity: AgendaEntity): AgendaSchedule {
            return AgendaSchedule(
                    id = scheduleAgendaEntity.id,
                    timeSpan = TimeSpan(
                            scheduleAgendaEntity.start, scheduleAgendaEntity.end
                    ),
                    attendee = scheduleAgendaEntity.attendee,
                    cancelReason = scheduleAgendaEntity.cancelReason,
                    data = when(scheduleAgendaEntity.data) {
                        is AgendaAppointmentData -> AppointmentScheduleData(
                                services = scheduleAgendaEntity.data.services.map { AppointmentService(it) }
                        )
                        is AgendaBasicData -> BasicScheduleData(
                                title = scheduleAgendaEntity.data.title,
                                description = scheduleAgendaEntity.data.description,
                        )
                    },
                    createdAt = scheduleAgendaEntity.createdAt
                )
        }
    }
}