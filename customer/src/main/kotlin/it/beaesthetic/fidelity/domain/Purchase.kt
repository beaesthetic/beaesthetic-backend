package it.beaesthetic.fidelity.domain

import java.time.Instant

sealed interface Purchase {
    val at: Instant
    val amount: Double
}

data class TreatmentPurchase(
    override val at: Instant,
    override val amount: Double,
    val treatment: FidelityTreatment
) : Purchase

enum class FidelityTreatment {
    SOLARIUM
}
