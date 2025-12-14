package it.beaesthetic.appointment.agenda.config

import com.fasterxml.jackson.databind.ObjectMapper
import io.quarkus.jackson.ObjectMapperCustomizer
import io.quarkus.runtime.annotations.RegisterForReflection
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventId
import it.beaesthetic.appointment.agenda.domain.notification.NotificationId
import it.beaesthetic.appointment.agenda.domain.notification.NotificationType
import it.beaesthetic.appointment.agenda.domain.notification.PendingNotification
import it.beaesthetic.appointment.agenda.infra.NotificationTypeMixin
import jakarta.inject.Singleton

@Singleton
@RegisterForReflection(
    targets =
        [
            PendingNotification::class,
            NotificationId::class,
            AgendaEventId::class,
            NotificationType::class,
            NotificationType.Reminder::class,
            NotificationType.Confirmation::class,
        ],
    registerFullHierarchy = true,
)
class RegisterCustomJackson : ObjectMapperCustomizer {
    override fun customize(objectMapper: ObjectMapper) {
        objectMapper.addMixIn(NotificationType::class.java, NotificationTypeMixin::class.java)
    }
}
