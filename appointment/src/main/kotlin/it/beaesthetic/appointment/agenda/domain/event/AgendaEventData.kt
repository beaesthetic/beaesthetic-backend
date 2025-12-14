package it.beaesthetic.appointment.agenda.domain.event

sealed interface AgendaEventData

data class BasicEventData(val title: String, val description: String?) : AgendaEventData

data class AppointmentService(val name: String)

data class AppointmentEventData(val services: List<AppointmentService>) : AgendaEventData
