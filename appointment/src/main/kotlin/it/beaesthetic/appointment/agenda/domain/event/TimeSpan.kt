package it.beaesthetic.appointment.agenda.domain.event

import java.time.Instant
import java.time.OffsetDateTime

data class TimeSpan(val start: Instant, val end: Instant) {
    init {
        require(start == end || start.isBefore(end)) { "start ($start) must be before ($end)" }
    }

    companion object {
        fun fromOffsetDateTime(start: OffsetDateTime, end: OffsetDateTime): TimeSpan {
            return TimeSpan(start.toInstant(), end.toInstant())
        }
    }
}
