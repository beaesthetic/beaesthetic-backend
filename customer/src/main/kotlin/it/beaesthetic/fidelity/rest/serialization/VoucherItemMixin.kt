package it.beaesthetic.fidelity.rest.serialization

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonSubTypes
import com.fasterxml.jackson.annotation.JsonTypeInfo
import it.beaesthetic.fidelity.generated.api.model.FreeVoucherDto
import it.beaesthetic.fidelity.generated.api.model.TreatmentDiscountVoucherDto

@JsonIgnoreProperties(value = ["type"], allowSetters = true)
@JsonTypeInfo(
    use = JsonTypeInfo.Id.NAME,
    property = "type",
    include = JsonTypeInfo.As.EXTERNAL_PROPERTY,
    visible = true
)
@JsonSubTypes(
    value =
        [
//            JsonSubTypes.Type(value = TreatmentDiscountVoucherDto::class, name = "foods"),
            JsonSubTypes.Type(value = FreeVoucherDto::class, name = "FreeVoucher"),
        ]
)
interface VoucherItemMixin
