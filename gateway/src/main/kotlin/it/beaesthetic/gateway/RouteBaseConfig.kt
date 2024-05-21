package it.beaesthetic.gateway

import org.springframework.boot.context.properties.ConfigurationProperties

@ConfigurationProperties(prefix = "routes.services")
data class RouteBaseConfig(
    val appointmentUrl: String,
    val customerUrl: String,
    val insightsUrl: String,
    val enableStrip: Boolean = false
)