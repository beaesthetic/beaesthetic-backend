package it.beaesthetic.notification.domain

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

class NotificationTest {

    private val notificationId = "notif-123"
    private val title = "Welcome"
    private val content = "Thank you for joining us"

    @Test
    fun `should create Notification with factory method`() {
        val channel = Email("user@example.com")

        val notification = Notification.of(notificationId, title, content, channel)

        assertEquals(notificationId, notification.id)
        assertEquals(title, notification.title)
        assertEquals(content, notification.content)
        assertEquals(channel, notification.channel)
        assertFalse(notification.isSent)
        assertFalse(notification.isSentConfirmed)
        assertNull(notification.channelMetadata)
        assertNotNull(notification.createdAt)
    }

    @Test
    fun `should emit NotificationCreated event when notification is created`() {
        val channel = Sms("+393331234567")

        val notification = Notification.of(notificationId, title, content, channel)

        val events = notification.events
        assertEquals(1, events.size)

        val event = events.first() as NotificationCreated
        assertEquals(notificationId, event.notificationId)
    }

    @Test
    fun `should mark notification as sent with channel metadata`() {
        val channel = Email("user@example.com")
        val notification = Notification.of(notificationId, title, content, channel)
        notification.clearEvents()

        val metadata = ChannelMetadata("provider-msg-123")
        val updated = notification.markSendWith(metadata)

        assertTrue(updated.isSent)
        assertEquals(metadata, updated.channelMetadata)
        assertFalse(updated.isSentConfirmed)
    }

    @Test
    fun `should emit NotificationSent event when marking as sent`() {
        val channel = Sms("+393331234567")
        val notification = Notification.of(notificationId, title, content, channel)
        notification.clearEvents()

        val metadata = ChannelMetadata("provider-msg-123")
        val updated = notification.markSendWith(metadata)

        val events = updated.events
        assertEquals(1, events.size)

        val event = events.first() as NotificationSent
        assertEquals(notificationId, event.notificationId)
    }

    @Test
    fun `should not emit event when marking already sent notification as sent`() {
        val channel = Email("user@example.com")
        val notification = Notification.of(notificationId, title, content, channel)
        val metadata = ChannelMetadata("provider-msg-123")
        val sent = notification.markSendWith(metadata)
        sent.clearEvents()

        val updated = sent.markSendWith(metadata)

        assertTrue(updated.isSent)
        assertTrue(updated.events.isEmpty())
    }

    @Test
    fun `should confirm notification send`() {
        val channel = Email("user@example.com")
        val notification = Notification.of(notificationId, title, content, channel)
        val metadata = ChannelMetadata("provider-msg-123")
        val sent = notification.markSendWith(metadata)
        sent.clearEvents()

        val confirmed = sent.confirmSend()

        assertTrue(confirmed.isSentConfirmed)
        assertTrue(confirmed.isSent)
    }

    @Test
    fun `should emit NotificationSentConfirmed event when confirming send`() {
        val channel = Sms("+393331234567")
        val notification = Notification.of(notificationId, title, content, channel)
        val metadata = ChannelMetadata("provider-msg-123")
        val sent = notification.markSendWith(metadata)
        sent.clearEvents()

        val confirmed = sent.confirmSend()

        val events = confirmed.events
        assertEquals(1, events.size)

        val event = events.first() as NotificationSentConfirmed
        assertEquals(notificationId, event.notificationId)
    }

    @Test
    fun `should not emit event when confirming already confirmed notification`() {
        val channel = Email("user@example.com")
        val notification = Notification.of(notificationId, title, content, channel)
        val metadata = ChannelMetadata("provider-msg-123")
        val sent = notification.markSendWith(metadata)
        val confirmed = sent.confirmSend()
        confirmed.clearEvents()

        val updated = confirmed.confirmSend()

        assertTrue(updated.isSentConfirmed)
        assertTrue(updated.events.isEmpty())
    }

    @Test
    fun `should support Email channel`() {
        val channel = Email("user@example.com")
        val notification = Notification.of(notificationId, title, content, channel)

        assertTrue(notification.isChannelSupported(Email::class.java))
        assertFalse(notification.isChannelSupported(Sms::class.java))
        assertFalse(notification.isChannelSupported(WhatsApp::class.java))
    }

    @Test
    fun `should support SMS channel`() {
        val channel = Sms("+393331234567")
        val notification = Notification.of(notificationId, title, content, channel)

        assertTrue(notification.isChannelSupported(Sms::class.java))
        assertFalse(notification.isChannelSupported(Email::class.java))
        assertFalse(notification.isChannelSupported(WhatsApp::class.java))
    }

    @Test
    fun `should support WhatsApp channel`() {
        val channel = WhatsApp("+393331234567")
        val notification = Notification.of(notificationId, title, content, channel)

        assertTrue(notification.isChannelSupported(WhatsApp::class.java))
        assertFalse(notification.isChannelSupported(Email::class.java))
        assertFalse(notification.isChannelSupported(Sms::class.java))
    }

    @Test
    fun `should preserve title and content through state transitions`() {
        val channel = Email("user@example.com")
        val notification = Notification.of(notificationId, title, content, channel)
        val metadata = ChannelMetadata("provider-msg-123")

        val sent = notification.markSendWith(metadata)
        val confirmed = sent.confirmSend()

        assertEquals(title, confirmed.title)
        assertEquals(content, confirmed.content)
        assertEquals(channel, confirmed.channel)
        assertEquals(notificationId, confirmed.id)
    }

    @Test
    fun `should handle complete notification lifecycle`() {
        val channel = Sms("+393331234567")

        // Create
        val notification = Notification.of(notificationId, title, content, channel)
        assertFalse(notification.isSent)
        assertFalse(notification.isSentConfirmed)
        assertNull(notification.channelMetadata)

        // Send
        val metadata = ChannelMetadata("provider-msg-123")
        val sent = notification.markSendWith(metadata)
        assertTrue(sent.isSent)
        assertFalse(sent.isSentConfirmed)
        assertEquals(metadata, sent.channelMetadata)

        // Confirm
        val confirmed = sent.confirmSend()
        assertTrue(confirmed.isSent)
        assertTrue(confirmed.isSentConfirmed)
        assertEquals(metadata, confirmed.channelMetadata)
    }

    @Test
    fun `should create notification with different channel types`() {
        val emailChannel = Email("user@example.com")
        val smsChannel = Sms("+393331234567")
        val whatsAppChannel = WhatsApp("+393331234567")

        val emailNotif = Notification.of("email-1", title, content, emailChannel)
        val smsNotif = Notification.of("sms-1", title, content, smsChannel)
        val whatsAppNotif = Notification.of("whatsapp-1", title, content, whatsAppChannel)

        assertTrue(emailNotif.channel is Email)
        assertTrue(smsNotif.channel is Sms)
        assertTrue(whatsAppNotif.channel is WhatsApp)
    }
}
