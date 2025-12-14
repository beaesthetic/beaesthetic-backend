package it.beaesthetic.appointment.agenda.port.rest

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonSubTypes
import com.fasterxml.jackson.annotation.JsonTypeInfo
import it.beaesthetic.appointment.agenda.generated.api.model.AppointmentEventDto
import it.beaesthetic.appointment.agenda.generated.api.model.GenericEventDto

@JsonIgnoreProperties(value = ["type"], allowSetters = true)
@JsonTypeInfo(
    use = JsonTypeInfo.Id.NAME,
    property = "eventType",
    include = JsonTypeInfo.As.EXTERNAL_PROPERTY,
    visible = true,
)
@JsonSubTypes(
    value =
        [
            JsonSubTypes.Type(value = GenericEventDto::class, name = "event"),
            JsonSubTypes.Type(value = AppointmentEventDto::class, name = "appointment"),
        ]
)
interface CreateAgendaActivityMixin
