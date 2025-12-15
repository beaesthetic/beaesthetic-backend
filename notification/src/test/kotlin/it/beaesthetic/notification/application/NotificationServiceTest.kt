package it.beaesthetic.notification.application

import it.beaesthetic.notification.domain.*
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking
import org.mockito.kotlin.*

class NotificationServiceTest {

    private val notificationRepository: NotificationRepository = mock()
    private val notificationProvider: NotificationProvider = mock()
    private val service = NotificationService(notificationRepository, notificationProvider)

    private val title = "Welcome"
    private val content = "Thank you for joining us"

    @Test
    fun `should create notification with Email channel`(): Unit = runBlocking {
        // Given
        val channel = Email("user@example.com")

        doAnswer { invocation -> invocation.getArgument<Notification>(0) }
            .whenever(notificationRepository)
            .save(any())

        // When
        val result = service.createNotification(title, content, channel)

        // Then
        assertTrue(result.isSuccess)
        val notification = result.getOrThrow()
        assertEquals(title, notification.title)
        assertEquals(content, notification.content)
        assertEquals(channel, notification.channel)
        assertNotNull(notification.id)

        verify(notificationRepository).save(any())
    }

    @Test
    fun `should create notification with SMS channel`(): Unit = runBlocking {
        // Given
        val channel = Sms("+393331234567")

        doAnswer { invocation -> invocation.getArgument<Notification>(0) }
            .whenever(notificationRepository)
            .save(any())

        // When
        val result = service.createNotification(title, content, channel)

        // Then
        assertTrue(result.isSuccess)
        val notification = result.getOrThrow()
        assertEquals(channel, notification.channel)

        verify(notificationRepository).save(any())
    }

    @Test
    fun `should emit NotificationCreated event when creating notification`(): Unit = runBlocking {
        // Given
        val channel = Email("user@example.com")

        doAnswer { invocation -> invocation.getArgument<Notification>(0) }
            .whenever(notificationRepository)
            .save(any())

        // When
        val result = service.createNotification(title, content, channel)

        // Then
        assertTrue(result.isSuccess)
        val notification = result.getOrThrow()
        val events = notification.events
        assertEquals(1, events.size)
        assertTrue(events.first() is NotificationCreated)
    }

    @Test
    fun `should send notification successfully when provider is supported`(): Unit = runBlocking {
        // Given
        val notificationId = "notif-123"
        val channel = Sms("+393331234567")
        val notification = Notification.of(notificationId, title, content, channel)
        val metadata = ChannelMetadata("provider-msg-123")

        whenever(notificationRepository.findById(notificationId)).thenReturn(notification)
        whenever(notificationProvider.isSupported(notification)).thenReturn(true)
        whenever(notificationProvider.send(notification)).thenReturn(Result.success(metadata))
        doAnswer { invocation -> invocation.getArgument<Notification>(0) }
            .whenever(notificationRepository)
            .save(any())

        // When
        val result = service.sendNotification(notificationId)

        // Then
        assertTrue(result.isSuccess)
        val sent = result.getOrThrow()
        assertTrue(sent.isSent)
        assertEquals(metadata, sent.channelMetadata)

        verify(notificationRepository).findById(notificationId)
        verify(notificationProvider).isSupported(notification)
        verify(notificationProvider).send(notification)
        verify(notificationRepository).save(any())
    }

    @Test
    fun `should fail to send notification when notification not found`(): Unit = runBlocking {
        // Given
        val notificationId = "non-existent"

        whenever(notificationRepository.findById(notificationId)).thenReturn(null)

        // When
        val result = service.sendNotification(notificationId)

        // Then
        assertTrue(result.isFailure)
        assertEquals("Notification $notificationId not found", result.exceptionOrNull()?.message)

        verify(notificationRepository).findById(notificationId)
        verify(notificationProvider, never()).isSupported(any())
        verify(notificationProvider, never()).send(any())
        verify(notificationRepository, never()).save(any())
    }

    @Test
    fun `should fail to send notification when provider is not supported`(): Unit = runBlocking {
        // Given
        val notificationId = "notif-123"
        val channel = WhatsApp("+393331234567")
        val notification = Notification.of(notificationId, title, content, channel)

        whenever(notificationRepository.findById(notificationId)).thenReturn(notification)
        whenever(notificationProvider.isSupported(notification)).thenReturn(false)

        // When
        val result = service.sendNotification(notificationId)

        // Then
        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull() is NoSuchElementException)

        verify(notificationRepository).findById(notificationId)
        verify(notificationProvider).isSupported(notification)
        verify(notificationProvider, never()).send(any())
        verify(notificationRepository, never()).save(any())
    }

    @Test
    fun `should fail to send notification when provider send fails`(): Unit = runBlocking {
        // Given
        val notificationId = "notif-123"
        val channel = Sms("+393331234567")
        val notification = Notification.of(notificationId, title, content, channel)
        val error = RuntimeException("Provider error")

        whenever(notificationRepository.findById(notificationId)).thenReturn(notification)
        whenever(notificationProvider.isSupported(notification)).thenReturn(true)
        whenever(notificationProvider.send(notification)).thenReturn(Result.failure(error))

        // When
        val result = service.sendNotification(notificationId)

        // Then
        assertTrue(result.isFailure)
        assertEquals(error, result.exceptionOrNull())

        verify(notificationRepository).findById(notificationId)
        verify(notificationProvider).send(notification)
        verify(notificationRepository, never()).save(any())
    }

    @Test
    fun `should emit NotificationSent event when sending notification`(): Unit = runBlocking {
        // Given
        val notificationId = "notif-123"
        val channel = Email("user@example.com")
        val notification = Notification.of(notificationId, title, content, channel)
        notification.clearEvents()
        val metadata = ChannelMetadata("provider-msg-123")

        whenever(notificationRepository.findById(notificationId)).thenReturn(notification)
        whenever(notificationProvider.isSupported(notification)).thenReturn(true)
        whenever(notificationProvider.send(notification)).thenReturn(Result.success(metadata))
        doAnswer { invocation -> invocation.getArgument<Notification>(0) }
            .whenever(notificationRepository)
            .save(any())

        // When
        val result = service.sendNotification(notificationId)

        // Then
        assertTrue(result.isSuccess)
        val sent = result.getOrThrow()
        val events = sent.events
        assertEquals(1, events.size)
        assertTrue(events.first() is NotificationSent)
    }

    @Test
    fun `should confirm notification sent successfully`(): Unit = runBlocking {
        // Given
        val notificationId = "notif-123"
        val channel = Sms("+393331234567")
        val metadata = ChannelMetadata("provider-msg-123")
        val notification =
            Notification.of(notificationId, title, content, channel).markSendWith(metadata)

        whenever(notificationRepository.findById(notificationId)).thenReturn(notification)
        doAnswer { invocation -> invocation.getArgument<Notification>(0) }
            .whenever(notificationRepository)
            .save(any())

        // When
        val result = service.confirmNotificationSent(notificationId)

        // Then
        assertTrue(result.isSuccess)
        val confirmed = result.getOrThrow()
        assertTrue(confirmed.isSent)
        assertTrue(confirmed.isSentConfirmed)

        verify(notificationRepository).findById(notificationId)
        verify(notificationRepository).save(any())
    }

    @Test
    fun `should fail to confirm notification when notification not found`(): Unit = runBlocking {
        // Given
        val notificationId = "non-existent"

        whenever(notificationRepository.findById(notificationId)).thenReturn(null)

        // When
        val result = service.confirmNotificationSent(notificationId)

        // Then
        assertTrue(result.isFailure)
        assertEquals("Notification $notificationId not found", result.exceptionOrNull()?.message)

        verify(notificationRepository).findById(notificationId)
        verify(notificationRepository, never()).save(any())
    }

    @Test
    fun `should emit NotificationSentConfirmed event when confirming`(): Unit = runBlocking {
        // Given
        val notificationId = "notif-123"
        val channel = Email("user@example.com")
        val metadata = ChannelMetadata("provider-msg-123")
        val notification =
            Notification.of(notificationId, title, content, channel).markSendWith(metadata)
        notification.clearEvents()

        whenever(notificationRepository.findById(notificationId)).thenReturn(notification)
        doAnswer { invocation -> invocation.getArgument<Notification>(0) }
            .whenever(notificationRepository)
            .save(any())

        // When
        val result = service.confirmNotificationSent(notificationId)

        // Then
        assertTrue(result.isSuccess)
        val confirmed = result.getOrThrow()
        val events = confirmed.events
        assertEquals(1, events.size)
        assertTrue(events.first() is NotificationSentConfirmed)
    }

    @Test
    fun `should propagate repository save failure when creating notification`(): Unit =
        runBlocking {
            // Given
            val channel = Email("user@example.com")
            val error = RuntimeException("Database error")

            doReturn(Result.failure<Notification>(error))
                .whenever(notificationRepository)
                .save(any())

            // When
            val result = service.createNotification(title, content, channel)

            // Then
            assertTrue(result.isFailure)
            assertEquals(error, result.exceptionOrNull())
        }

    @Test
    fun `should propagate repository save failure when sending notification`(): Unit = runBlocking {
        // Given
        val notificationId = "notif-123"
        val channel = Sms("+393331234567")
        val notification = Notification.of(notificationId, title, content, channel)
        val metadata = ChannelMetadata("provider-msg-123")
        val error = RuntimeException("Database error")

        whenever(notificationRepository.findById(notificationId)).thenReturn(notification)
        whenever(notificationProvider.isSupported(notification)).thenReturn(true)
        whenever(notificationProvider.send(notification)).thenReturn(Result.success(metadata))
        doReturn(Result.failure<Notification>(error)).whenever(notificationRepository).save(any())

        // When
        val result = service.sendNotification(notificationId)

        // Then
        assertTrue(result.isFailure)
        assertEquals(error, result.exceptionOrNull())
    }

    @Test
    fun `should propagate repository save failure when confirming notification`(): Unit =
        runBlocking {
            // Given
            val notificationId = "notif-123"
            val channel = Email("user@example.com")
            val metadata = ChannelMetadata("provider-msg-123")
            val notification =
                Notification.of(notificationId, title, content, channel).markSendWith(metadata)
            val error = RuntimeException("Database error")

            whenever(notificationRepository.findById(notificationId)).thenReturn(notification)
            doReturn(Result.failure<Notification>(error))
                .whenever(notificationRepository)
                .save(any())

            // When
            val result = service.confirmNotificationSent(notificationId)

            // Then
            assertTrue(result.isFailure)
            assertEquals(error, result.exceptionOrNull())
        }
}
