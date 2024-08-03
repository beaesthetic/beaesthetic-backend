package it.beaesthetic.notification.infra.providers

import it.beaesthetic.notification.domain.ChannelMetadata
import it.beaesthetic.notification.domain.Notification
import it.beaesthetic.notification.domain.NotificationProvider

class WhatsappNotificationProvider : NotificationProvider {
    override suspend fun isSupported(notification: Notification): Boolean {
        TODO("Not yet implemented")
    }

    override suspend fun send(notification: Notification): Result<ChannelMetadata> {
        TODO("Not yet implemented")
    }
}
