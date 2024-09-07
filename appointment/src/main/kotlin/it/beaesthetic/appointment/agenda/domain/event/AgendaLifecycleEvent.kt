package it.beaesthetic.appointment.agenda.domain.event

sealed interface AgendaLifecycleEvent

data class AgendaEventScheduled(val agendaEvent: AgendaEvent) : AgendaLifecycleEvent

data class AgendaEventRescheduled(val agendaEvent: AgendaEvent) : AgendaLifecycleEvent

data class AgendaEventDeleted(val agendaEvent: AgendaEvent) : AgendaLifecycleEvent
