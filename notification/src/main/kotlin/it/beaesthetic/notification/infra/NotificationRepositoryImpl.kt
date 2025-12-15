package it.beaesthetic.notification.infra

import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.smallrye.mutiny.coroutines.awaitSuspending
import io.vertx.mutiny.core.eventbus.EventBus
import it.beaesthetic.notification.EventConfiguration
import it.beaesthetic.notification.common.DomainEventRegistryDelegate
import it.beaesthetic.notification.domain.Notification
import it.beaesthetic.notification.domain.NotificationRepository
import jakarta.enterprise.context.ApplicationScoped
import java.time.Instant

@ApplicationScoped
class PanacheNotificationRepository : ReactivePanacheMongoRepository<NotificationEntity>

@ApplicationScoped
class NotificationRepositoryImpl(
    private val panacheNotificationRepository: PanacheNotificationRepository,
    private val eventBus: EventBus,
) : NotificationRepository {

    override suspend fun findById(notificationId: String): Notification? {
        return panacheNotificationRepository
            .find("_id", notificationId)
            .firstResult()
            .map { if (it != null) EntityMapper.toDomain(it) else null }
            .awaitSuspending()
    }

    override suspend fun save(notification: Notification): Result<Notification> = runCatching {
        panacheNotificationRepository
            .persistOrUpdate(EntityMapper.toEntity(notification))
            .awaitSuspending()
            .also {
                notification.events.forEach {
                    eventBus.publish(EventConfiguration.INTERNAL_DOMAIN_EVENTS_TOPIC, it)
                }
            }
        notification
    }

    private object EntityMapper {
        fun toEntity(notification: Notification): NotificationEntity =
            NotificationEntity(
                id = notification.id,
                title = notification.title,
                content = notification.content,
                isSent = notification.isSent,
                isSentConfirmed = notification.isSentConfirmed,
                channel = notification.channel,
                channelData = notification.channelMetadata,
                createdAt = notification.createdAt,
                updatedAt = Instant.now(),
            )

        fun toDomain(entity: NotificationEntity): Notification =
            Notification(
                id = entity.id,
                title = entity.title,
                content = entity.content,
                isSent = entity.isSent,
                isSentConfirmed = entity.isSentConfirmed,
                channel = entity.channel,
                channelMetadata = entity.channelData,
                createdAt = entity.createdAt,
                domainEventRegistry = DomainEventRegistryDelegate(),
            )
    }
}
