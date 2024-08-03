package it.beaesthetic.notification.domain

interface NotificationProvider {
    suspend fun isSupported(notification: Notification): Boolean
    suspend fun send(notification: Notification): Result<ChannelMetadata>
}

class CompoundNotificationProvider(private val providers: List<NotificationProvider>) :
    NotificationProvider {

    override suspend fun isSupported(notification: Notification): Boolean {
        return findProviderFor(notification) != null
    }

    override suspend fun send(notification: Notification): Result<ChannelMetadata> {
        val provider = findProviderFor(notification)
        return provider?.send(notification)
            ?: Result.failure(NoSuchElementException("No provider found for $notification"))
    }

    private suspend fun findProviderFor(notification: Notification): NotificationProvider? {
        return providers.firstOrNull { it.isSupported(notification) }
    }
}
