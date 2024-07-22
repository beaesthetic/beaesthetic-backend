package it.beaesthetic.fidelity

import it.beaesthetic.fidelity.application.FidelityCardService
import it.beaesthetic.fidelity.domain.FidelityCardRepository
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces

@Dependent
class DependencyConfiguration {

    @Produces
    fun fidelityCardService(
        fidelityCardRepository: FidelityCardRepository
    ): FidelityCardService {
        return FidelityCardService(fidelityCardRepository)
    }
}