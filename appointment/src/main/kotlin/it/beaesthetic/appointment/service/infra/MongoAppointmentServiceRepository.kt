package it.beaesthetic.appointment.service.infra

import com.mongodb.client.model.IndexModel
import com.mongodb.client.model.IndexOptions
import com.mongodb.client.model.Indexes
import io.quarkus.mongodb.FindOptions
import io.quarkus.mongodb.panache.common.MongoEntity
import io.quarkus.mongodb.panache.kotlin.PanacheMongoEntityBase
import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.quarkus.runtime.Startup
import io.quarkus.runtime.annotations.RegisterForReflection
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.service.common.SearchGram
import it.beaesthetic.appointment.service.domain.AppointmentService
import it.beaesthetic.appointment.service.domain.AppointmentServiceRepository
import it.beaesthetic.appointment.service.domain.Color
import jakarta.enterprise.context.ApplicationScoped
import kotlinx.coroutines.runBlocking
import org.bson.Document
import org.bson.codecs.pojo.annotations.BsonCreator
import org.bson.codecs.pojo.annotations.BsonId
import org.bson.codecs.pojo.annotations.BsonProperty
import org.bson.conversions.Bson

@RegisterForReflection
@MongoEntity(collection = "services")
data class AppointmentServiceEntity
@BsonCreator
constructor(
    @BsonId val id: String,
    @BsonProperty("name") val name: String,
    @BsonProperty("price") val price: Double,
    @BsonProperty("tags") val tags: List<String>,
    @BsonProperty("colorHex") val colorHex: String?,
    @BsonProperty("searchGrams") val searchGrams: String? = null,
) : PanacheMongoEntityBase()

@ApplicationScoped
class PanacheAppointmentRepository : ReactivePanacheMongoRepository<AppointmentServiceEntity> {
    @Startup
    fun initializeIndexes() = runBlocking {
        mongoCollection()
            .createIndexes(
                listOf(
                    IndexModel(Indexes.text("searchGrams")),
                    IndexModel(Indexes.ascending("name"), IndexOptions().unique(true)),
                )
            )
            .awaitSuspending()
    }
}

@ApplicationScoped
class MongoAppointmentServiceRepository(
    private val panacheAppointmentRepository: PanacheAppointmentRepository
) : AppointmentServiceRepository {

    override suspend fun searchByName(searchTerm: String, limit: Int): List<AppointmentService> {
        val fieldsToProject = listOf("name", "price", "tags", "colorHex")
        val filter: Bson = Document("\$text", Document("\$search", searchTerm))
        val project: Bson =
            Document("score", Document("\$meta", "textScore")).apply {
                fieldsToProject.forEach { field -> append(field, 1L) }
            }
        val sort: Bson = Document("score", Document("\$meta", "textScore"))
        return panacheAppointmentRepository
            .mongoCollection()
            .find(FindOptions().filter(filter).projection(project).sort(sort).limit(limit))
            .map { Mapper.toDomain(it) }
            .collect()
            .asList()
            .awaitSuspending()
    }

    override suspend fun findAll(): List<AppointmentService> {
        return panacheAppointmentRepository.findAll().list().awaitSuspending().map {
            Mapper.toDomain(it)
        }
    }

    override suspend fun findById(id: String): AppointmentService? {
        return panacheAppointmentRepository
            .find("_id", id)
            .firstResult()
            .map { e -> e?.let { Mapper.toDomain(it) } }
            .awaitSuspending()
    }

    override suspend fun save(appointmentService: AppointmentService): Result<AppointmentService> {
        val searchTerms =
            appointmentService.tags.flatMap { SearchGram.Default.nGrams(it, 2) } +
                SearchGram.Default.nGrams(appointmentService.name.trim().lowercase(), 2)
        val entity =
            Mapper.toEntity(appointmentService)
                .copy(searchGrams = searchTerms.toSet().joinToString(" "))
        return runCatching {
            panacheAppointmentRepository.persistOrUpdate(entity).awaitSuspending().let {
                Mapper.toDomain(entity)
            }
        }
    }

    override suspend fun delete(id: String): Result<String> = runCatching {
        panacheAppointmentRepository.delete("_id", id).awaitSuspending()
        id
    }

    private object Mapper {
        fun toDomain(entity: AppointmentServiceEntity): AppointmentService =
            AppointmentService(
                id = entity.id,
                name = entity.name,
                price = entity.price,
                tags = entity.tags.toSet(),
                color = entity.colorHex?.let { Color(it) }
            )

        fun toEntity(service: AppointmentService): AppointmentServiceEntity =
            AppointmentServiceEntity(
                id = service.id,
                name = service.name,
                price = service.price,
                tags = service.tags.toList(),
                colorHex = service.color?.hex
            )
    }
}
