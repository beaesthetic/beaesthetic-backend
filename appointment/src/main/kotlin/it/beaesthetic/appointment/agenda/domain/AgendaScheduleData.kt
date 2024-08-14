package it.beaesthetic.appointment.agenda.domain

sealed interface AgendaScheduleData
data class BasicScheduleData(
    val title: String,
    val description: String?
) : AgendaScheduleData


data class AppointmentService(val name: String)

data class AppointmentScheduleData(
    val services: List<AppointmentService>,
) : AgendaScheduleData