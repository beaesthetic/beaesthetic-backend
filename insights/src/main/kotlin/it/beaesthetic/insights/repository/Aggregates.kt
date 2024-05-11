package it.beaesthetic.insights.repository

import it.beaesthetic.insights.service.InsightService
import org.bson.BsonArray
import org.bson.BsonDocument

object Aggregates {
    fun templatizeQuery(
        aggregateQueryPath: String,
        vararg params: Pair<String, Any>
    ) = templatizeQuery(aggregateQueryPath, params.toMap())

    fun templatizeQuery(
        aggregateQueryPath: String,
        templateParams: Map<String, Any>
    ): List<BsonDocument> {
        return InsightService::class.java.getResourceAsStream(aggregateQueryPath)
            ?.bufferedReader()
            ?.readText()
            ?.let {
                templateParams.entries.fold(it) { query, (key, value) ->
                    query.replace("{${key}}", value.toString(), ignoreCase = true)
                }
            }
            ?.let { BsonArray.parse(it) }
            ?.map { it.asDocument() }
            ?: emptyList()
    }
}