package it.beaesthetic.appointment.agenda.infra.mongo

import io.quarkus.mongodb.panache.common.MongoEntity
import io.quarkus.mongodb.panache.kotlin.PanacheMongoEntityBase
import io.quarkus.runtime.annotations.RegisterForReflection
import it.beaesthetic.appointment.agenda.domain.event.CancelReason
import java.time.Instant
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonDiscriminator
import org.bson.codecs.pojo.annotations.BsonId
import org.bson.codecs.pojo.annotations.BsonProperty

@RegisterForReflection
@MongoEntity(collection = "agenda")
data class AgendaEntity
@BsonCreator
constructor(
    @BsonId val id: String,
    @BsonProperty("start") val start: Instant,
    @BsonProperty("end") val end: Instant,
    @BsonProperty("attendee") val attendee: AttendeeEntity,
    @BsonProperty("data") val data: AgendaScheduleData,
    @BsonProperty("cancelReason") val cancelReason: CancelReason?,
    @BsonProperty("remindBeforeSeconds") val remindBeforeSeconds: Int,
    @BsonProperty("reminderStatus") val reminderStatus: String,
    @BsonProperty("isCancelled") val isCancelled: Boolean,
    @BsonProperty("version") val version: Long,
    @BsonProperty("createdAt") val createdAt: Instant,
    @BsonProperty("updatedAt") val updatedAt: Instant
) : PanacheMongoEntityBase()

@RegisterForReflection
data class AttendeeEntity
@BsonCreator
constructor(
    @get:BsonProperty("id") @param:BsonProperty("id") val id: String,
    @BsonProperty("displayName") val displayName: String
)

@RegisterForReflection @BsonDiscriminator(key = "type") sealed interface AgendaScheduleData

@RegisterForReflection
@BsonDiscriminator(key = "type", value = "appointment")
data class AgendaAppointmentData
@BsonCreator
constructor(
    @BsonProperty("services") val services: List<String>,
) : AgendaScheduleData

@RegisterForReflection
@BsonDiscriminator(key = "type", value = "event")
data class AgendaBasicData
@BsonCreator
constructor(
    @BsonProperty("title") val title: String,
    @BsonProperty("description") val description: String,
) : AgendaScheduleData
