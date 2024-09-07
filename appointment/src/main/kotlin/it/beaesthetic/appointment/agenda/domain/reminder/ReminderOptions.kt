package it.beaesthetic.appointment.agenda.domain.reminder

import java.time.Duration

data class ReminderOptions(
    val wantRecap: Boolean = false,
    val triggerBefore: Duration = Duration.ZERO
)
