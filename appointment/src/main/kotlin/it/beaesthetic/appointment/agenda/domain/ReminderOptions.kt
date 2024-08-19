package it.beaesthetic.appointment.agenda.domain

import java.time.Duration

data class ReminderOptions(val triggerBefore: Duration = Duration.ZERO)
