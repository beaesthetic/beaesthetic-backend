package it.beaesthetic.fidelity.application.data

import io.quarkus.runtime.annotations.RegisterForReflection
import it.beaesthetic.fidelity.domain.Voucher

@RegisterForReflection
data class FidelityCardReadModel(
    val id: String,
    val customer: CustomerInfo,
    val vouchers: List<Voucher>,
    val solariumPurchases: Int,
)

@RegisterForReflection
data class CustomerInfo(
    val id: String,
    val name: String,
    val surname: String,
    val email: String,
    val phone: String,
)
