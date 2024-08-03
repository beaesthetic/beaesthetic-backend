package it.beaesthetic.notification.infra

import com.fasterxml.jackson.annotation.JsonSubTypes
import com.fasterxml.jackson.annotation.JsonTypeInfo
import it.beaesthetic.notification.domain.NotificationCreated
import it.beaesthetic.notification.domain.NotificationSent
import it.beaesthetic.notification.domain.NotificationSentConfirmed
import it.beaesthetic.notification.infra.NotificationEventMixin.Companion.NOTIFICATION_CREATED_TYPE
import it.beaesthetic.notification.infra.NotificationEventMixin.Companion.NOTIFICATION_SENT_CONFIRMED_TYPE
import it.beaesthetic.notification.infra.NotificationEventMixin.Companion.NOTIFICATION_SENT_TYPE

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, property = "type")
@JsonSubTypes(
    JsonSubTypes.Type(value = NotificationCreated::class, name = NOTIFICATION_CREATED_TYPE),
    JsonSubTypes.Type(value = NotificationSent::class, name = NOTIFICATION_SENT_TYPE),
    JsonSubTypes.Type(value = NotificationSentConfirmed::class, name = NOTIFICATION_SENT_CONFIRMED_TYPE)
)
interface NotificationEventMixin {
    companion object {
        const val NOTIFICATION_CREATED_TYPE = "notification.created"
        const val NOTIFICATION_SENT_TYPE = "notification.sent"
        const val NOTIFICATION_SENT_CONFIRMED_TYPE = "notification.sent.confirmed"
    }
}
