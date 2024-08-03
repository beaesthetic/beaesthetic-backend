package it.beaesthetic.notification.domain

sealed interface NotificationEvent

data class NotificationCreated(val notificationId: String) : NotificationEvent
data class NotificationSent(val notificationId: String) : NotificationEvent
data class NotificationSentConfirmed(val notificationId: String) : NotificationEvent