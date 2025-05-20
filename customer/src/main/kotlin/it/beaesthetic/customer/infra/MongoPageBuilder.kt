package it.beaesthetic.customer.infra

import com.mongodb.client.model.Collation
import com.mongodb.client.model.CollationStrength
import io.quarkus.mongodb.FindOptions
import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.quarkus.panache.common.Sort
import io.smallrye.mutiny.coroutines.awaitSuspending
import java.util.Base64
import org.bson.Document

data class Page<T>(val items: List<T>, val lastItemToken: String?)

class MongoPageBuilder<K, T : Any>(
    private val repository: ReactivePanacheMongoRepository<T>,
    private val pageSize: Int,
    private val sortFields: List<String>,
    private val direction: Sort.Direction,
    private val idExtractor: (T) -> K,
) {

    suspend fun paginate(pageToken: String?): Page<T> {
        val filter =
            when {
                pageToken != null -> generateCursorFilter(pageToken)
                else -> Document()
            }
        var sort = generateSortDocument()
        var collation =
            Collation.builder()
                .locale("en")
                .caseLevel(false)
                .collationStrength(CollationStrength.SECONDARY)
                .build()
        var items =
            repository
                .mongoCollection()
                .find(FindOptions().filter(filter).sort(sort).limit(pageSize).collation(collation))
                .collect()
                .asList()
                .awaitSuspending()

        return Page(
            items,
            lastItemToken =
                if (items.size < pageSize) null
                else getNextPageToken(idExtractor(items.last()), items.last()),
        )
    }

    private fun generateCursorFilter(pageToken: String) =
        generateCursorFilter(decodePageToken(pageToken))

    private fun generateCursorFilter(cursorValues: List<String>): Document {
        require(sortFields.size == cursorValues.size) { "Cursor and sort field mismatch" }
        val op = if (direction == Sort.Direction.Ascending) "\$gt" else "\$lt"

        val orConditions = mutableListOf<Document>()
        for (i in sortFields.indices) {
            val andCondition = Document()
            for (j in 0 until i) {
                andCondition.append(sortFields[j], cursorValues[j])
            }
            andCondition.append(sortFields[i], Document(op, cursorValues[i]))
            orConditions.add(andCondition)
        }

        return Document("\$or", orConditions)
    }

    private fun generateSortDocument(): Document {
        val sortDir = if (direction == Sort.Direction.Ascending) 1 else -1
        return sortFields.fold(Document()) { doc, field -> doc.append(field, sortDir) }
    }

    private fun <T> getNextPageToken(id: K, item: T): String {
        val values =
            sortFields.map { field ->
                when (field) {
                    "_id" -> id.toString()
                    else ->
                        item!!::class
                            .java
                            .getDeclaredField(field)
                            .apply { isAccessible = true }
                            .get(item)
                            ?.toString() ?: ""
                }
            }
        return encodePageToken(values)
    }

    private fun encodePageToken(values: List<String>): String {
        return Base64.getEncoder().encodeToString(values.joinToString("__").encodeToByteArray())
    }

    private fun decodePageToken(pageToken: String): List<String> {
        return String(Base64.getDecoder().decode(pageToken)).split("__")
    }
}
