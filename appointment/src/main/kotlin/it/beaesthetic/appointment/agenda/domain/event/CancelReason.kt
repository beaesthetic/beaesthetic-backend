package it.beaesthetic.appointment.agenda.domain.event

sealed interface CancelReason

data object CustomerCancel : CancelReason

data object NoReason : CancelReason
