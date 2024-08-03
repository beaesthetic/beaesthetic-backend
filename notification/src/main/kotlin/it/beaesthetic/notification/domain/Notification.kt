package it.beaesthetic.notification.domain

import it.beaesthetic.notification.common.DomainEventRegistry
import it.beaesthetic.notification.common.DomainEventRegistryDelegate
import java.time.Instant

data class Notification(
    val id: String,
    val title: String,
    val content: String,
    val isSent: Boolean,
    val isSentConfirmed: Boolean,
    val channel: Channel,
    val channelMetadata: ChannelMetadata?,
    val createdAt: Instant,
    private val domainEventRegistry: DomainEventRegistry<NotificationEvent>
) : DomainEventRegistry<NotificationEvent> by domainEventRegistry {

    companion object {
        fun of(id: String, title: String, content: String, channel: Channel): Notification =
            Notification(
                    id = id,
                    title = title,
                    content = content,
                    channel = channel,
                    isSent = false,
                    isSentConfirmed = false,
                    channelMetadata = null,
                    createdAt = Instant.now(),
                    domainEventRegistry = DomainEventRegistryDelegate()
                )
                .apply { addEvent(NotificationCreated(id)) }
    }

    fun <T : Channel> isChannelSupported(klazz: Class<T>): Boolean {
        return klazz.isAssignableFrom(channel::class.java)
    }

    fun markSendWith(channelMetadata: ChannelMetadata): Notification {
        if (!isSent) {
            addEvent(NotificationSent(id))
            return copy(isSent = true, channelMetadata = channelMetadata)
        }
        return this
    }

    fun confirmSend(): Notification {
        if (!isSentConfirmed) {
            addEvent(NotificationSentConfirmed(id))
            return copy(isSentConfirmed = true)
        }
        return this
    }
}
