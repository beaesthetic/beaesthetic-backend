package it.beaesthetic.fidelity.application

import it.beaesthetic.fidelity.application.data.FidelityCardReadModel
import it.beaesthetic.fidelity.domain.CustomerId

interface FidelityCardReadService {
    suspend fun findAll(): List<FidelityCardReadModel>

    suspend fun findById(id: String): FidelityCardReadModel?

    suspend fun findByCustomerId(customerId: CustomerId): FidelityCardReadModel?
}
