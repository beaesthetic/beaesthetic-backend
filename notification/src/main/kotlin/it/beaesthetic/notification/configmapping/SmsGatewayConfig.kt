package it.beaesthetic.notification.configmapping

import io.smallrye.config.ConfigMapping

@ConfigMapping(prefix = "smsGateway")
interface SmsGatewayConfig {
    fun senderNumber(): String
}