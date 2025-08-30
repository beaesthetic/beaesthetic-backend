package it.beaesthetic.fidelity.rest

import it.beaesthetic.fidelity.application.data.FidelityCardReadModel
import it.beaesthetic.fidelity.domain.Voucher
import it.beaesthetic.fidelity.generated.api.model.CustomerDto
import it.beaesthetic.fidelity.generated.api.model.FidelityCardResponseDto
import it.beaesthetic.fidelity.generated.api.model.FreeVoucherDto
import it.beaesthetic.fidelity.generated.api.model.SupportedVoucherTreatmentDto
import java.time.ZoneOffset
import java.util.*

object ResourceMapper {

    fun Voucher.toResource() =
        FreeVoucherDto(
            id = UUID.fromString(id.value),
            issuedAt = createdAt.atOffset(ZoneOffset.UTC),
            treatment = SupportedVoucherTreatmentDto.valueOf(treatment.name),
            isUsed = isUsed,
        )

    fun FidelityCardReadModel.toResource(): FidelityCardResponseDto {
        return FidelityCardResponseDto(
            id = UUID.fromString(this.id),
            customer =
                CustomerDto(
                    id = customer.id,
                    name = customer.name ?: "",
                    surname = customer.surname ?: "",
                    phone = customer.phone,
                    email = customer.email,
                ),
            solariumPurchases = solariumPurchases,
            vouchers =
                vouchers.map {
                    FreeVoucherDto(
                        id = UUID.fromString(it.id),
                        issuedAt = it.createdAt.atOffset(ZoneOffset.UTC),
                        treatment = SupportedVoucherTreatmentDto.valueOf(it.treatment.name),
                        isUsed = it.isUsed,
                    )
                },
        )
    }
}
