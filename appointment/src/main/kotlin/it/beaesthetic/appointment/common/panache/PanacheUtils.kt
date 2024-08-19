package it.beaesthetic.appointment.common.panache

import com.mongodb.client.result.UpdateResult
import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.smallrye.mutiny.Uni
import org.bson.BsonDocument
import org.bson.BsonDocumentWriter
import org.bson.Document
import org.bson.codecs.EncoderContext
import org.bson.conversions.Bson

object PanacheUtils {

    fun <T : Any> ReactivePanacheMongoRepository<T>.updateOne(
        filter: Bson,
        entity: T
    ): Uni<UpdateResult> {
        val bsonDocument = BsonDocument()
        val codec = mongoCollection().codecRegistry.get(entity.javaClass)
        codec.encode(BsonDocumentWriter(bsonDocument), entity, EncoderContext.builder().build())
        return mongoCollection().updateOne(filter, Document("\$set", bsonDocument))
    }
}
