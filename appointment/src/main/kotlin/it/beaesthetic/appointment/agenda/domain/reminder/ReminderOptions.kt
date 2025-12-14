package it.beaesthetic.appointment.agenda.domain.reminder

import java.time.Duration

data class ReminderOptions(
    val sendBefore: Duration,
    val noSendThreshold: Duration,
    val immediateSendThreshold: Duration,
)
