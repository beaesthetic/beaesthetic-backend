package it.beaesthetic.customer.infra

import com.fasterxml.jackson.databind.ObjectMapper
import io.quarkus.mongodb.reactive.ReactiveMongoDatabase
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.common.DomainEventRegistry
import org.bson.BsonDocument
import org.bson.BsonValue
import org.bson.Document

class OutboxRepository<E>(
    private val mongoDatabase: ReactiveMongoDatabase,
    private val objectMapper: ObjectMapper,
    private val outboxCollectionName: String
) {
    private val collection by lazy { mongoDatabase.getCollection(outboxCollectionName) }

    private suspend fun save(event: E): BsonValue? {
        val rawEvent = objectMapper.writeValueAsString(event)
        return collection.insertOne(Document.parse(rawEvent)).awaitSuspending().insertedId
    }

    suspend fun save(domainEventRegistry: DomainEventRegistry<E>) {
        domainEventRegistry.events.forEach {
            val id = save(it)
            collection.deleteOne(BsonDocument("_id", id)).awaitSuspending()
        }
        domainEventRegistry.clearEvents()
    }
}
