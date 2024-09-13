package it.beaesthetic.appointment.agenda.domain

import java.time.Instant

fun interface Clock {
    fun now(): Instant

    companion object {
        fun default(): Clock = Clock { Instant.now() }
    }
}
