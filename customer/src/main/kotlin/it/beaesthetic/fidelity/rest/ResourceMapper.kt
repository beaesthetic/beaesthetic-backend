package it.beaesthetic.fidelity.rest

import it.beaesthetic.fidelity.domain.FidelityCard
import it.beaesthetic.fidelity.domain.FidelityTreatment
import it.beaesthetic.fidelity.domain.Voucher
import it.beaesthetic.fidelity.generated.api.model.FidelityCardResponseDto
import it.beaesthetic.fidelity.generated.api.model.FreeVoucherDto
import it.beaesthetic.fidelity.generated.api.model.SupportedVoucherTreatmentDto
import java.time.ZoneOffset
import java.util.*

object ResourceMapper {
    fun FidelityCard.toResource() = FidelityCardResponseDto(
        id = UUID.fromString(this.id),
        customerId = UUID.fromString(this.customerId.id),
        solariumPurchases = purchasesOf(FidelityTreatment.SOLARIUM),
        vouchers = this.availableVouchers.map {
            FreeVoucherDto(
                id = UUID.fromString(it.id.value),
                issuedAt = it.createdAt.atOffset(ZoneOffset.UTC),
                treatment = SupportedVoucherTreatmentDto.valueOf(it.treatment.name),
                isUsed = it.isUsed,
            )
        }
    )

    fun Voucher.toResource() = FreeVoucherDto(
        id = UUID.fromString(id.value),
        issuedAt = createdAt.atOffset(ZoneOffset.UTC),
        treatment = SupportedVoucherTreatmentDto.valueOf(treatment.name),
        isUsed = isUsed,
    )
}