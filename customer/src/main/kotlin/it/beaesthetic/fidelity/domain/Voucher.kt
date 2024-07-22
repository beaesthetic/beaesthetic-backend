package it.beaesthetic.fidelity.domain

import java.time.Instant
import java.util.UUID

@JvmInline value class VoucherId(val value: String)

data class Voucher(
    val id: VoucherId,
    val createdAt: Instant,
    val isUsed: Boolean,
    val treatment: FidelityTreatment
) {

    companion object {
        fun ofTreatment(treatment: FidelityTreatment) =
            Voucher(
                id = VoucherId(UUID.randomUUID().toString()),
                createdAt = Instant.now(),
                isUsed = false,
                treatment = treatment
            )
    }
}
