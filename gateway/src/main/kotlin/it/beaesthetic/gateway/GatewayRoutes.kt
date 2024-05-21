package it.beaesthetic.gateway

import it.beaesthetic.gateway.appointment.EnrichWithCustomerDataFilter
import it.beaesthetic.gateway.filters.RemoveIfNoneMatchHeader
import org.springframework.cloud.gateway.filter.GlobalFilter
import org.springframework.cloud.gateway.filter.factory.rewrite.ModifyResponseBodyGatewayFilterFactory
import org.springframework.cloud.gateway.route.RouteLocator
import org.springframework.cloud.gateway.route.builder.RouteLocatorBuilder
import org.springframework.cloud.gateway.route.builder.filters
import org.springframework.cloud.gateway.route.builder.routes
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.http.HttpMethod

@Configuration
class GatewayRoutes {

    @Bean
    fun enrichWithCustomerData(modifyResponse: ModifyResponseBodyGatewayFilterFactory) =
        EnrichWithCustomerDataFilter(modifyResponse)

    @Bean
    fun removeIfNoneMatch(): GlobalFilter {
        return RemoveIfNoneMatchHeader()
    }

    @Bean
    fun gatewayRouteLocator(
        builder: RouteLocatorBuilder,
        routeBaseConfig: RouteBaseConfig,
        enrichWithCustomerData: EnrichWithCustomerDataFilter
    ): RouteLocator {
        val enrichConfig = EnrichWithCustomerDataFilter.Config(
            customerIdField = "customerId",
            customerField = "customer",
            customerUrl = "${routeBaseConfig.customerUrl}/customer/{customerId}",
            excludeProperties = setOf("note"),
            forwardHeader = false
        )

        return builder.routes {
            route("appointment-service") {
                uri(routeBaseConfig.appointmentUrl)
                path("/appointment-service/**")
                order(0)
                filters {
                    if (routeBaseConfig.enableStrip) stripPrefix(1)
                }
            }

            route("customer-get-fidelity-card") {
                uri(routeBaseConfig.customerUrl)
                path("/customer-service/fidelity-card/**")
                    .and()
                    .method(HttpMethod.GET, HttpMethod.POST)
                order(0)
                filters {
                    if (routeBaseConfig.enableStrip) stripPrefix(1)
                    filter(enrichWithCustomerData.apply(enrichConfig), -1)
                }
            }

            route("customer-get-gift-card") {
                uri(routeBaseConfig.customerUrl)
                path("/customer-service/gift-card/**")
                    .and()
                    .method(HttpMethod.GET, HttpMethod.POST)
                order(0)
                filters {
                    if (routeBaseConfig.enableStrip) stripPrefix(1)
                    filter(enrichWithCustomerData.apply(enrichConfig), -1)
                }
            }

            route("customer-get-fidelity-cards") {
                uri(routeBaseConfig.customerUrl)
                path("/customer-service/fidelity-cards/**")
                    .and()
                    .method(HttpMethod.GET, HttpMethod.POST)
                order(0)
                filters {
                    if (routeBaseConfig.enableStrip) stripPrefix(1)
                    filter(enrichWithCustomerData.apply(enrichConfig), -1)
                }
            }

            route("customer-get-gift-cards") {
                uri(routeBaseConfig.customerUrl)
                path("/customer-service/gift-cards/**")
                    .and()
                    .method(HttpMethod.GET, HttpMethod.POST)
                order(0)
                filters {
                    if (routeBaseConfig.enableStrip) stripPrefix(1)
                    filter(enrichWithCustomerData.apply(enrichConfig), -1)
                }
            }

            route("customer-get-wallets") {
                uri(routeBaseConfig.customerUrl)
                path("/customer-service/wallets/**")
                    .and()
                    .method(HttpMethod.GET, HttpMethod.POST)
                order(0)
                filters {
                    if (routeBaseConfig.enableStrip) stripPrefix(1)
                    filter(enrichWithCustomerData.apply(enrichConfig), -1)
                }
            }

            route("customer-get-wallets") {
                uri(routeBaseConfig.customerUrl)
                path("/customer-service/wallets/**")
                    .and()
                    .method(HttpMethod.GET, HttpMethod.POST)
                order(0)
                filters {
                    if (routeBaseConfig.enableStrip) stripPrefix(1)
                    filter(enrichWithCustomerData.apply(enrichConfig), -1)
                }
            }

            route("customer-service") {
                uri(routeBaseConfig.customerUrl)
                path("/customer-service/**")
                order(1)
                filters {
                    if (routeBaseConfig.enableStrip) stripPrefix(1)
                }
            }

            route("insights-service") {
                uri(routeBaseConfig.insightsUrl)
                path("/insights-service/**")
                order(1)
                filters {
                    if (routeBaseConfig.enableStrip) stripPrefix(1)
                }
            }
        }
    }
}