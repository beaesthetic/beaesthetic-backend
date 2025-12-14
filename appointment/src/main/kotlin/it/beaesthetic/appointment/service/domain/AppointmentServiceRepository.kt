package it.beaesthetic.appointment.service.domain

interface AppointmentServiceRepository {
    suspend fun searchByName(searchTerm: String, limit: Int): List<AppointmentService>

    suspend fun findAll(): List<AppointmentService>

    suspend fun findById(id: String): AppointmentService?

    suspend fun save(appointmentService: AppointmentService): Result<AppointmentService>

    suspend fun delete(id: String): Result<String>
}
