package it.beaesthetic.notification.domain

import kotlinx.coroutines.runBlocking
import org.mockito.kotlin.*
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class CompoundNotificationProviderTest {

    private val smsProvider: NotificationProvider = mock()
    private val emailProvider: NotificationProvider = mock()
    private val whatsAppProvider: NotificationProvider = mock()

    private val title = "Welcome"
    private val content = "Thank you for joining"

    @Test
    fun `should find provider that supports SMS channel`(): Unit = runBlocking {
        // Given
        val providers = listOf(emailProvider, smsProvider, whatsAppProvider)
        val compound = CompoundNotificationProvider(providers)
        val notification = Notification.of("notif-1", title, content, Sms("+393331234567"))

        whenever(emailProvider.isSupported(notification)).thenReturn(false)
        whenever(smsProvider.isSupported(notification)).thenReturn(true)

        // When
        val isSupported = compound.isSupported(notification)

        // Then
        assertTrue(isSupported)
        verify(emailProvider).isSupported(notification)
        verify(smsProvider).isSupported(notification)
        verify(whatsAppProvider, never()).isSupported(any())
    }

    @Test
    fun `should find provider that supports Email channel`(): Unit = runBlocking {
        // Given
        val providers = listOf(smsProvider, emailProvider, whatsAppProvider)
        val compound = CompoundNotificationProvider(providers)
        val notification = Notification.of("notif-1", title, content, Email("user@example.com"))

        whenever(smsProvider.isSupported(notification)).thenReturn(false)
        whenever(emailProvider.isSupported(notification)).thenReturn(true)

        // When
        val isSupported = compound.isSupported(notification)

        // Then
        assertTrue(isSupported)
        verify(smsProvider).isSupported(notification)
        verify(emailProvider).isSupported(notification)
        verify(whatsAppProvider, never()).isSupported(any())
    }

    @Test
    fun `should return false when no provider supports channel`(): Unit = runBlocking {
        // Given
        val providers = listOf(smsProvider, emailProvider)
        val compound = CompoundNotificationProvider(providers)
        val notification = Notification.of("notif-1", title, content, WhatsApp("+393331234567"))

        whenever(smsProvider.isSupported(notification)).thenReturn(false)
        whenever(emailProvider.isSupported(notification)).thenReturn(false)

        // When
        val isSupported = compound.isSupported(notification)

        // Then
        assertFalse(isSupported)
        verify(smsProvider).isSupported(notification)
        verify(emailProvider).isSupported(notification)
    }

    @Test
    fun `should send notification using first supporting provider`(): Unit = runBlocking {
        // Given
        val providers = listOf(emailProvider, smsProvider)
        val compound = CompoundNotificationProvider(providers)
        val notification = Notification.of("notif-1", title, content, Sms("+393331234567"))
        val metadata = ChannelMetadata("provider-msg-123")

        whenever(emailProvider.isSupported(notification)).thenReturn(false)
        whenever(smsProvider.isSupported(notification)).thenReturn(true)
        whenever(smsProvider.send(notification)).thenReturn(Result.success(metadata))

        // When
        val result = compound.send(notification)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(metadata, result.getOrThrow())

        verify(emailProvider).isSupported(notification)
        verify(smsProvider).isSupported(notification)
        verify(smsProvider).send(notification)
        verify(emailProvider, never()).send(any())
    }

    @Test
    fun `should fail when no provider supports notification`(): Unit = runBlocking {
        // Given
        val providers = listOf(smsProvider, emailProvider)
        val compound = CompoundNotificationProvider(providers)
        val notification = Notification.of("notif-1", title, content, WhatsApp("+393331234567"))

        whenever(smsProvider.isSupported(notification)).thenReturn(false)
        whenever(emailProvider.isSupported(notification)).thenReturn(false)

        // When
        val result = compound.send(notification)

        // Then
        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull() is NoSuchElementException)

        verify(smsProvider).isSupported(notification)
        verify(emailProvider).isSupported(notification)
        verify(smsProvider, never()).send(any())
        verify(emailProvider, never()).send(any())
    }

    @Test
    fun `should propagate provider send failure`(): Unit = runBlocking {
        // Given
        val providers = listOf(smsProvider)
        val compound = CompoundNotificationProvider(providers)
        val notification = Notification.of("notif-1", title, content, Sms("+393331234567"))
        val error = RuntimeException("Provider API error")

        whenever(smsProvider.isSupported(notification)).thenReturn(true)
        whenever(smsProvider.send(notification)).thenReturn(Result.failure(error))

        // When
        val result = compound.send(notification)

        // Then
        assertTrue(result.isFailure)
        assertEquals(error, result.exceptionOrNull())

        verify(smsProvider).send(notification)
    }

    @Test
    fun `should use first provider when multiple providers support same channel`(): Unit = runBlocking {
        // Given
        val smsProvider1: NotificationProvider = mock()
        val smsProvider2: NotificationProvider = mock()
        val providers = listOf(smsProvider1, smsProvider2)
        val compound = CompoundNotificationProvider(providers)
        val notification = Notification.of("notif-1", title, content, Sms("+393331234567"))
        val metadata = ChannelMetadata("provider1-msg-123")

        whenever(smsProvider1.isSupported(notification)).thenReturn(true)
        whenever(smsProvider1.send(notification)).thenReturn(Result.success(metadata))

        // When
        val result = compound.send(notification)

        // Then
        assertTrue(result.isSuccess)
        assertEquals(metadata, result.getOrThrow())

        verify(smsProvider1).isSupported(notification)
        verify(smsProvider1).send(notification)
        verify(smsProvider2, never()).isSupported(any())
        verify(smsProvider2, never()).send(any())
    }

    @Test
    fun `should handle empty provider list`(): Unit = runBlocking {
        // Given
        val compound = CompoundNotificationProvider(emptyList())
        val notification = Notification.of("notif-1", title, content, Sms("+393331234567"))

        // When
        val isSupported = compound.isSupported(notification)
        val sendResult = compound.send(notification)

        // Then
        assertFalse(isSupported)
        assertTrue(sendResult.isFailure)
        assertTrue(sendResult.exceptionOrNull() is NoSuchElementException)
    }

    @Test
    fun `should check all providers in order until supported one is found`(): Unit = runBlocking {
        // Given
        val provider1: NotificationProvider = mock()
        val provider2: NotificationProvider = mock()
        val provider3: NotificationProvider = mock()
        val providers = listOf(provider1, provider2, provider3)
        val compound = CompoundNotificationProvider(providers)
        val notification = Notification.of("notif-1", title, content, Email("user@example.com"))

        whenever(provider1.isSupported(notification)).thenReturn(false)
        whenever(provider2.isSupported(notification)).thenReturn(false)
        whenever(provider3.isSupported(notification)).thenReturn(true)

        // When
        val isSupported = compound.isSupported(notification)

        // Then
        assertTrue(isSupported)
        verify(provider1).isSupported(notification)
        verify(provider2).isSupported(notification)
        verify(provider3).isSupported(notification)
    }

    @Test
    fun `should stop checking providers after first supported one is found`(): Unit = runBlocking {
        // Given
        val provider1: NotificationProvider = mock()
        val provider2: NotificationProvider = mock()
        val provider3: NotificationProvider = mock()
        val providers = listOf(provider1, provider2, provider3)
        val compound = CompoundNotificationProvider(providers)
        val notification = Notification.of("notif-1", title, content, Sms("+393331234567"))
        val metadata = ChannelMetadata("provider2-msg-123")

        whenever(provider1.isSupported(notification)).thenReturn(false)
        whenever(provider2.isSupported(notification)).thenReturn(true)
        whenever(provider2.send(notification)).thenReturn(Result.success(metadata))

        // When
        compound.send(notification)

        // Then
        verify(provider1).isSupported(notification)
        verify(provider2).isSupported(notification)
        verify(provider3, never()).isSupported(any())
        verify(provider2).send(notification)
    }
}
