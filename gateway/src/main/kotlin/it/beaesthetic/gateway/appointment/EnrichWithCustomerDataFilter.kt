package it.beaesthetic.gateway.appointment;

import com.fasterxml.jackson.databind.JsonNode
import com.fasterxml.jackson.databind.node.ArrayNode
import com.fasterxml.jackson.databind.node.ObjectNode
import it.beaesthetic.gateway.utils.JsonNodeExtensions.excludeFields
import it.beaesthetic.gateway.utils.ModifyResponseBodyGatewayExtensions.setJsonRewriteFunction
import it.beaesthetic.gateway.utils.WebClientExtensions.headers
import org.jboss.logging.Logger
import org.springframework.cloud.gateway.filter.GatewayFilter
import org.springframework.cloud.gateway.filter.GatewayFilterChain
import org.springframework.cloud.gateway.filter.NettyWriteResponseFilter
import org.springframework.cloud.gateway.filter.factory.AbstractGatewayFilterFactory
import org.springframework.cloud.gateway.filter.factory.rewrite.ModifyResponseBodyGatewayFilterFactory
import org.springframework.core.Ordered
import org.springframework.http.HttpHeaders
import org.springframework.web.reactive.function.client.WebClient
import org.springframework.web.reactive.function.client.bodyToMono
import org.springframework.web.server.ServerWebExchange
import reactor.core.publisher.Flux
import reactor.core.publisher.Mono

class EnrichWithCustomerDataFilter(
    private val modifyResponseBodyGatewayFilterFactory: ModifyResponseBodyGatewayFilterFactory
) : AbstractGatewayFilterFactory<EnrichWithCustomerDataFilter.Config>() {

    data class Config(
        val customerIdField: String,
        val customerField: String,
        val customerUrl: String,
        val excludeProperties: Set<String> = emptySet(),
        val forwardHeader: Boolean = false,
    )

    companion object {
        private val logger = Logger.getLogger(EnrichWithCustomerDataFilter::class.java)
        private val successRangeResponse = 200..300
    }

    private val webClient: WebClient = WebClient.create()

    override fun apply(config: Config): GatewayFilter {
        val modifyResponseBodyFilterFactoryConfig = ModifyResponseBodyGatewayFilterFactory.Config().apply {
            setJsonRewriteFunction { exchange, list ->
                compositeResponse(config, list, if (config.forwardHeader) exchange.request.headers else HttpHeaders())
            }
        }

        return object : GatewayFilter, Ordered {
            override fun filter(exchange: ServerWebExchange, chain: GatewayFilterChain): Mono<Void> {
                return when (exchange.response.statusCode?.value()) {
                    in successRangeResponse -> modifyResponseBodyGatewayFilterFactory
                        .apply(modifyResponseBodyFilterFactoryConfig)
                        .filter(exchange, chain)

                    else -> chain.filter(exchange)
                }
            }

            override fun getOrder(): Int = NettyWriteResponseFilter.WRITE_RESPONSE_FILTER_ORDER - 1
        }
    }

    private fun compositeResponse(
        config: Config,
        json: JsonNode,
        httpHeaders: HttpHeaders
    ): Mono<JsonNode> = when (json) {
        is ObjectNode -> replaceWithCompositeCall(json, config, httpHeaders)
        is ArrayNode -> Flux.fromIterable(Iterable { json.elements() })
            .filter { it.isObject }
            .flatMap { replaceWithCompositeCall(it as ObjectNode, config, httpHeaders) }
            .onErrorResume {
                logger.warn("Error getting composite data, ignoring it", it)
                Mono.empty()
            }
            .collect({ json }, { array, e -> array })

        else -> Mono.just(json)
    }

    private fun replaceWithCompositeCall(
        json: ObjectNode,
        config: Config,
        httpHeaders: HttpHeaders
    ): Mono<JsonNode> {
        return json.get(config.customerIdField)
            ?.textValue()
            ?.let { getCompositeData(it, config, httpHeaders) }
            ?.map {
                json.apply {
                    putIfAbsent(config.customerField, it)
                    remove(config.customerIdField)
                }
            }
            ?: Mono.just(json)
    }

    private fun getCompositeData(
        customerId: String,
        config: Config,
        httpHeaders: HttpHeaders
    ): Mono<JsonNode> = webClient.get()
        .uri(config.customerUrl, customerId)
        .apply { if (config.forwardHeader) headers(httpHeaders) }
        .retrieve()
        .bodyToMono<JsonNode>()
        .map { it.excludeFields(config.excludeProperties) }
}
