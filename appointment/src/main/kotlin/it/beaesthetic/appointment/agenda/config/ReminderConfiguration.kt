package it.beaesthetic.appointment.agenda.config

import io.smallrye.config.ConfigMapping
import java.time.Duration

@ConfigMapping(prefix = "reminder")
interface ReminderConfiguration {
    fun triggerBefore(): Duration
    fun noSendThreshold(): Duration
    fun immediateSendThreshold(): Duration
}
