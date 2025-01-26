package it.beaesthetic.appointment.agenda.config

import io.smallrye.config.ConfigMapping

@ConfigMapping(prefix = "notification")
interface NotificationConfiguration {
    fun whitelist(): List<String>
}