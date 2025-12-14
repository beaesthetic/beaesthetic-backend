package it.beaesthetic.appointment.agenda.domain.event

import it.beaesthetic.appointment.agenda.domain.Clock
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import java.time.Duration
import java.time.Instant
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

class AgendaEventTest {

    private val fixedClock = Clock { Instant.parse("2024-01-15T12:00:00Z") }

    @Test
    fun `should create AgendaEvent with basic event data`() {
        // Given
        val id = AgendaEventId("event-123")
        val timeSpan =
            TimeSpan(Instant.parse("2024-01-20T10:00:00Z"), Instant.parse("2024-01-20T11:00:00Z"))
        val attendee = Attendee("customer-456", "John Doe")
        val data = BasicEventData("Meeting", "Important meeting")
        val reminderBefore = Duration.ofHours(1)

        // When
        val event = AgendaEvent.create(id, timeSpan, attendee, data, reminderBefore)

        // Then
        assertEquals(id, event.id)
        assertEquals(timeSpan, event.timeSpan)
        assertEquals(attendee, event.attendee)
        assertEquals(data, event.data)
        assertEquals(reminderBefore, event.remindBefore)
        assertNull(event.cancelReason)
        assertEquals(ReminderStatus.PENDING, event.reminder.status)
        assertEquals("Meeting", event.title)
    }

    @Test
    fun `should create AgendaEvent with appointment data`() {
        // Given
        val id = AgendaEventId("event-789")
        val timeSpan =
            TimeSpan(Instant.parse("2024-01-20T10:00:00Z"), Instant.parse("2024-01-20T11:00:00Z"))
        val attendee = Attendee("customer-456", "Jane Smith")
        val services = listOf(AppointmentService("Haircut"), AppointmentService("Massage"))
        val data = AppointmentEventData(services)
        val reminderBefore = Duration.ofMinutes(30)

        // When
        val event = AgendaEvent.create(id, timeSpan, attendee, data, reminderBefore)

        // Then
        assertEquals(data, event.data)
        assertEquals("Haircut,Massage", event.title)
    }

    @Test
    fun `should emit AgendaEventScheduled when event is created`() {
        // Given
        val id = AgendaEventId("event-123")
        val timeSpan =
            TimeSpan(Instant.parse("2024-01-20T10:00:00Z"), Instant.parse("2024-01-20T11:00:00Z"))
        val attendee = Attendee("customer-456", "John Doe")
        val data = BasicEventData("Meeting", "Test")
        val reminderBefore = Duration.ofHours(1)

        // When
        val event = AgendaEvent.create(id, timeSpan, attendee, data, reminderBefore)

        // Then
        val events = event.events
        assertEquals(1, events.size)
        assertTrue(events.first() is AgendaEventScheduled)
        assertEquals(event, (events.first() as AgendaEventScheduled).agendaEvent)
    }

    @Test
    fun `should reschedule event to new timespan`() {
        // Given
        val event = createTestEvent()
        val newTimeSpan =
            TimeSpan(Instant.parse("2024-01-21T14:00:00Z"), Instant.parse("2024-01-21T15:00:00Z"))

        // When
        val rescheduledEvent = event.reschedule(newTimeSpan)

        // Then
        assertEquals(newTimeSpan, rescheduledEvent.timeSpan)
        assertEquals(event.id, rescheduledEvent.id)
        assertEquals(event.attendee, rescheduledEvent.attendee)
        assertNull(rescheduledEvent.cancelReason)

        // Verify domain event
        val domainEvents = rescheduledEvent.events
        assertTrue(domainEvents.any { it is AgendaEventRescheduled })
    }

    @Test
    fun `should not reschedule already canceled event`() {
        // Given
        val event = createTestEvent().cancel(CustomerCancel)
        val newTimeSpan =
            TimeSpan(Instant.parse("2024-01-21T14:00:00Z"), Instant.parse("2024-01-21T15:00:00Z"))

        // When & Then
        val exception = assertThrows<IllegalArgumentException> { event.reschedule(newTimeSpan) }

        assertTrue(exception.message!!.contains("already canceled"))
    }

    @Test
    fun `should cancel event with reason`() {
        // Given
        val event = createTestEvent()

        // When
        val canceledEvent = event.cancel(CustomerCancel)

        // Then
        assertEquals(CustomerCancel, canceledEvent.cancelReason)
        assertEquals(event.id, canceledEvent.id)
        assertEquals(event.timeSpan, canceledEvent.timeSpan)

        // Verify domain event
        val domainEvents = canceledEvent.events
        assertTrue(domainEvents.any { it is AgendaEventDeleted })
    }

    @Test
    fun `should not cancel already canceled event`() {
        // Given
        val event = createTestEvent().cancel(CustomerCancel)

        // When & Then
        val exception = assertThrows<IllegalArgumentException> { event.cancel(NoReason) }

        assertTrue(exception.message!!.contains("already canceled"))
    }

    @Test
    fun `should update reminder status`() {
        // Given
        val event = createTestEvent()
        val newReminder = event.reminder.copy(status = ReminderStatus.SCHEDULED)

        // When
        val updatedEvent = event.updateReminder(newReminder)

        // Then
        assertEquals(ReminderStatus.SCHEDULED, updatedEvent.reminder.status)
        assertEquals(event.id, updatedEvent.id)
    }

    @Test
    fun `should track reminder as sent with timestamp`() {
        // Given
        val event = createTestEvent()

        // When
        val updatedEvent = event.trackReminderAsSent(fixedClock)

        // Then
        assertEquals(ReminderStatus.SENT, updatedEvent.reminder.status)
        assertNotNull(updatedEvent.reminder.sentAt)
        assertEquals(fixedClock.now(), updatedEvent.reminder.sentAt)
    }

    @Test
    fun `should preserve event identity across operations`() {
        // Given
        val event = createTestEvent()
        val eventId = event.id

        // When
        val rescheduled =
            event.reschedule(
                TimeSpan(
                    Instant.parse("2024-01-21T14:00:00Z"),
                    Instant.parse("2024-01-21T15:00:00Z"),
                )
            )
        val withUpdatedReminder =
            rescheduled.updateReminder(rescheduled.reminder.copy(status = ReminderStatus.SENT))

        // Then
        assertEquals(eventId, rescheduled.id)
        assertEquals(eventId, withUpdatedReminder.id)
    }

    private fun createTestEvent(): AgendaEvent {
        return AgendaEvent.create(
            id = AgendaEventId("test-event-123"),
            timeSpan =
                TimeSpan(
                    Instant.parse("2024-01-20T10:00:00Z"),
                    Instant.parse("2024-01-20T11:00:00Z"),
                ),
            attendee = Attendee("customer-123", "Test User"),
            data = BasicEventData("Test Event", "Description"),
            reminderBefore = Duration.ofHours(1),
        )
    }
}
