package it.beaesthetic.gateway.configuration

import org.springframework.cloud.gateway.filter.GlobalFilter
import org.springframework.cloud.gateway.filter.factory.DedupeResponseHeaderGatewayFilterFactory
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.core.annotation.Order


@Configuration
class CorsConfiguration {

    @Bean
    @Order(0)
    fun deduplicateCorsFilter(): GlobalFilter {
        val deduplication = DedupeResponseHeaderGatewayFilterFactory()
            .apply(
                DedupeResponseHeaderGatewayFilterFactory.Config()
                    .apply { name = "Access-Control-Allow-Credentials Access-Control-Allow-Origin" })
        return GlobalFilter { exchange, chain -> deduplication.filter(exchange, chain) }
    }
}