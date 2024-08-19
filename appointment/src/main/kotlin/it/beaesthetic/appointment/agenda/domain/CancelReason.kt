package it.beaesthetic.appointment.agenda.domain

sealed interface CancelReason

data object CustomerCancel : CancelReason

data object NoReason : CancelReason
