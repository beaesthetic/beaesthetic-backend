package it.beaesthetic.appointment.service.domain

data class Color(val hex: String)

data class AppointmentService(
    val id: String,
    val name: String,
    val price: Double,
    val tags: Set<String>,
    val color: Color?
)