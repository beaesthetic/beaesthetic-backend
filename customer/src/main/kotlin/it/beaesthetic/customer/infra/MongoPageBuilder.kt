package it.beaesthetic.customer.infra

import io.quarkus.panache.common.Sort
import java.util.Base64
import org.bson.Document

class MongoPageBuilder(
    private val sortFields: List<String>,
    private val direction: Sort.Direction
) {

    fun generateCursorFilter(pageToken: String) = generateCursorFilter(decodePageToken(pageToken))

    fun generateCursorFilter(cursorValues: List<String>): Document {
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

    fun generateSortDocument(): Document {
        val sortDir = if (direction == Sort.Direction.Ascending) 1 else -1
        return sortFields.fold(Document()) { doc, field -> doc.append(field, sortDir) }
    }

    fun <T> getNextPageToken(id: String, item: T): String {
        val values =
            sortFields.map { field ->
                when (field) {
                    "_id" -> id
                    else ->
                        item!!::class
                            .java
                            .getDeclaredField(field)
                            .apply { isAccessible = true }
                            .get(item)
                            .toString()
                }
            }
        return encodePageToken(values)
    }

    fun encodePageToken(values: List<String>): String {
        return Base64.getEncoder().encodeToString(values.joinToString("__").encodeToByteArray())
    }

    fun decodePageToken(pageToken: String): List<String> {
        return String(Base64.getDecoder().decode(pageToken)).split("__")
    }
}
