package it.beaesthetic.customer.infra

import com.fasterxml.jackson.databind.ObjectMapper
import io.quarkus.mongodb.reactive.ReactiveMongoDatabase
import io.smallrye.mutiny.Uni
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.common.DomainEventRegistry
import java.util.*
import org.bson.BsonDocument
import org.bson.BsonValue
import org.bson.Document

data class OutboxEvent<E>(val id: String, val eventType: String, val eventContent: E)

class OutboxRepository<E>(
    private val mongoDatabase: ReactiveMongoDatabase,
    private val objectMapper: ObjectMapper,
    private val outboxCollectionName: String
) {
    private val collection by lazy { mongoDatabase.getCollection(outboxCollectionName) }

    private suspend fun save(eventType: String, event: E): BsonValue? {
        val decoratedOutboxEvent =
            OutboxEvent(
                id = UUID.randomUUID().toString(),
                eventType = eventType,
                eventContent = event
            )
        val rawEvent = objectMapper.writeValueAsString(decoratedOutboxEvent)
        return collection
            .insertOne(Document.parse(rawEvent).apply { remove("_id") })
            .awaitSuspending()
            .insertedId
    }

    suspend fun save(domainEventRegistry: DomainEventRegistry<E>) {
        domainEventRegistry.events.forEach { (type, event) ->
            Uni.createFrom()
                .item(save(type, event))
                //                .onItem()
                //                .delayIt()
                //                .by(Duration.ofMillis(100))
                .flatMap { collection.deleteOne(BsonDocument("_id", it)) }
                .awaitSuspending()
        }
        domainEventRegistry.clearEvents()
    }
}
