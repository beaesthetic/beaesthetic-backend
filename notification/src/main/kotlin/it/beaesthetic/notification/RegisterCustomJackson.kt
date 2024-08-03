package it.beaesthetic.notification

import com.fasterxml.jackson.databind.ObjectMapper
import io.quarkus.jackson.ObjectMapperCustomizer
import it.beaesthetic.notification.domain.NotificationEvent
import it.beaesthetic.notification.infra.NotificationEventMixin
import jakarta.inject.Singleton

@Singleton
class RegisterCustomJackson : ObjectMapperCustomizer {
    override fun customize(objectMapper: ObjectMapper) {
        objectMapper.addMixIn(
            NotificationEvent::class.java, NotificationEventMixin::class.java
        )
    }
}