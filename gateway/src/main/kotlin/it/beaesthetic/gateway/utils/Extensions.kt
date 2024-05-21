package it.beaesthetic.gateway.utils

import com.fasterxml.jackson.databind.JsonNode
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.node.ObjectNode
import org.springframework.cloud.gateway.filter.factory.rewrite.ModifyResponseBodyGatewayFilterFactory
import org.springframework.cloud.gateway.filter.factory.rewrite.RewriteFunction
import org.springframework.http.HttpHeaders
import org.springframework.web.reactive.function.client.WebClient
import reactor.core.publisher.Mono

object WebClientExtensions {
    fun <S : WebClient.RequestHeadersSpec<S>> WebClient.RequestHeadersSpec<S>.headers(
        headers: HttpHeaders
    ): WebClient.RequestHeadersSpec<S> = this.apply {
        headers.toMap().forEach { (t, u) -> header(t, *u.toTypedArray()) }
    }
}

object ModifyResponseBodyGatewayExtensions {
    val mapper = ObjectMapper()
    inline fun <reified O> ModifyResponseBodyGatewayFilterFactory.Config.setJsonRewriteFunction(
        rewriteFunction: RewriteFunction<JsonNode, O>
    ) {
        setRewriteFunction(String::class.java, O::class.java) { exchange, body ->
            val parsedJson = mapper.readTree(body)
            Mono.from(rewriteFunction.apply(exchange, parsedJson))
                .map { it }
        }
    }
}

object JsonNodeExtensions {
    fun JsonNode.excludeFields(excludeFields: Set<String>, recursive: Boolean = false): JsonNode = when (this) {
        is ObjectNode -> remove(excludeFields.toList())
        else -> this
    }
}