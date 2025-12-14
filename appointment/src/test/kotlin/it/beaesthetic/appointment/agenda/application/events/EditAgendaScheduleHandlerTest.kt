package it.beaesthetic.appointment.agenda.application.events

import it.beaesthetic.appointment.agenda.domain.event.*
import it.beaesthetic.appointment.agenda.domain.reminder.Reminder
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import it.beaesthetic.appointment.common.DomainEventRegistry
import it.beaesthetic.appointment.common.OptimisticConcurrency.VersionedEntity
import java.time.Duration
import java.time.Instant
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.*

class EditAgendaScheduleHandlerTest {

    private lateinit var agendaRepository: AgendaRepository
    private lateinit var handler: EditAgendaScheduleHandler

    @BeforeEach
    fun setup() {
        agendaRepository = mock()
        handler = EditAgendaScheduleHandler(agendaRepository)
    }

    @Test
    fun `should edit basic event title and description`(): Unit = runBlocking {
        // Given
        val eventId = AgendaEventId("event-123")
        val existingEvent = createBasicEvent(eventId)
        val version = 5L

        val command =
            EditAgendaSchedule(
                scheduleId = eventId,
                timeSpan = null,
                services = null,
                title = "Updated Meeting",
                description = "Updated description",
            )

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        whenever(agendaRepository.saveEvent(any(), eq(version))).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isSuccess)
        val updatedEvent = result.getOrThrow()

        val data = updatedEvent.data as BasicEventData
        assertEquals("Updated Meeting", data.title)
        assertEquals("Updated description", data.description)
        assertEquals(existingEvent.timeSpan, updatedEvent.timeSpan)

        verify(agendaRepository).saveEvent(any(), eq(version))
    }

    @Test
    fun `should reschedule event to new timespan`(): Unit = runBlocking {
        // Given
        val eventId = AgendaEventId("event-123")
        val existingEvent = createBasicEvent(eventId)
        val version = 3L

        val newTimeSpan =
            TimeSpan(Instant.parse("2024-01-21T14:00:00Z"), Instant.parse("2024-01-21T15:00:00Z"))

        val command =
            EditAgendaSchedule(
                scheduleId = eventId,
                timeSpan = newTimeSpan,
                services = null,
                title = null,
                description = null,
            )

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        whenever(agendaRepository.saveEvent(any(), eq(version))).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isSuccess)
        val updatedEvent = result.getOrThrow()

        assertEquals(newTimeSpan, updatedEvent.timeSpan)
        verify(agendaRepository).saveEvent(any(), eq(version))
    }

    @Test
    fun `should edit appointment services`(): Unit = runBlocking {
        // Given
        val eventId = AgendaEventId("event-456")
        val existingEvent = createAppointmentEvent(eventId)
        val version = 2L

        val newServices = setOf(AppointmentService("Facial"), AppointmentService("Manicure"))

        val command =
            EditAgendaSchedule(
                scheduleId = eventId,
                timeSpan = null,
                services = newServices,
                title = null,
                description = null,
            )

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        whenever(agendaRepository.saveEvent(any(), eq(version))).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isSuccess)
        val updatedEvent = result.getOrThrow()

        val data = updatedEvent.data as AppointmentEventData
        assertEquals(newServices.toList(), data.services)
        assertEquals("Facial,Manicure", updatedEvent.title)

        verify(agendaRepository).saveEvent(any(), eq(version))
    }

    @Test
    fun `should edit both timespan and data together`() = runBlocking {
        // Given
        val eventId = AgendaEventId("event-789")
        val existingEvent = createBasicEvent(eventId)
        val version = 1L

        val newTimeSpan =
            TimeSpan(Instant.parse("2024-01-22T10:00:00Z"), Instant.parse("2024-01-22T11:00:00Z"))

        val command =
            EditAgendaSchedule(
                scheduleId = eventId,
                timeSpan = newTimeSpan,
                services = null,
                title = "Rescheduled Meeting",
                description = "New time and title",
            )

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        whenever(agendaRepository.saveEvent(any(), eq(version))).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isSuccess)
        val updatedEvent = result.getOrThrow()

        assertEquals(newTimeSpan, updatedEvent.timeSpan)
        val data = updatedEvent.data as BasicEventData
        assertEquals("Rescheduled Meeting", data.title)
        assertEquals("New time and title", data.description)
    }

    @Test
    fun `should fail when event not found`(): Unit = runBlocking {
        // Given
        val eventId = AgendaEventId("non-existent")
        val command =
            EditAgendaSchedule(
                scheduleId = eventId,
                timeSpan = null,
                services = null,
                title = "Test",
                description = null,
            )

        whenever(agendaRepository.findEvent(eventId)).thenReturn(null)

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull() is IllegalArgumentException)
        assertTrue(result.exceptionOrNull()!!.message!!.contains("Schedule not found"))

        verify(agendaRepository, never()).saveEvent(any(), any())
    }

    @Test
    fun `should fail when repository save fails`() = runBlocking {
        // Given
        val eventId = AgendaEventId("event-123")
        val existingEvent = createBasicEvent(eventId)
        val version = 2L

        val command =
            EditAgendaSchedule(
                scheduleId = eventId,
                timeSpan = null,
                services = null,
                title = "Updated",
                description = null,
            )

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        val repositoryError = RuntimeException("Database error")
        whenever(agendaRepository.saveEvent(any(), eq(version)))
            .thenReturn(Result.failure(repositoryError))

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isFailure)
        assertEquals(repositoryError, result.exceptionOrNull())
    }

    @Test
    fun `should preserve unchanged fields when editing`() = runBlocking {
        // Given
        val eventId = AgendaEventId("event-123")
        val existingEvent = createBasicEvent(eventId)
        val version = 1L

        val command =
            EditAgendaSchedule(
                scheduleId = eventId,
                timeSpan = null,
                services = null,
                title = "New Title", // Only changing title
                description = null, // Description should remain unchanged
            )

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))
        whenever(agendaRepository.saveEvent(any(), eq(version))).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result = handler.handle(command)

        // Then
        val updatedEvent = result.getOrThrow()
        val data = updatedEvent.data as BasicEventData

        assertEquals("New Title", data.title)
        assertEquals("Important meeting", data.description) // Original description preserved
        assertEquals(existingEvent.timeSpan, updatedEvent.timeSpan) // Original timespan preserved
        assertEquals(existingEvent.attendee, updatedEvent.attendee) // Attendee preserved
    }

    @Test
    fun `should use optimistic locking with correct version`(): Unit = runBlocking {
        // Given
        val eventId = AgendaEventId("event-123")
        val existingEvent = createBasicEvent(eventId)
        val version = 42L // Specific version number

        val command =
            EditAgendaSchedule(
                scheduleId = eventId,
                timeSpan = null,
                services = null,
                title = "Test",
                description = null,
            )

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        whenever(agendaRepository.saveEvent(any(), eq(version))).thenAnswer { invocation ->
            Result.success(invocation.getArgument<AgendaEvent>(0))
        }

        // When
        handler.handle(command)

        // Then
        verify(agendaRepository).saveEvent(any(), eq(42L))
    }

    private fun createBasicEvent(id: AgendaEventId): AgendaEvent {
        return AgendaEvent(
            id = id,
            timeSpan =
                TimeSpan(
                    Instant.parse("2024-01-20T10:00:00Z"),
                    Instant.parse("2024-01-20T11:00:00Z"),
                ),
            attendee = Attendee("customer-123", "John Doe"),
            cancelReason = null,
            data = BasicEventData("Meeting", "Important meeting"),
            reminder = Reminder(id, ReminderStatus.PENDING, null),
            remindBefore = Duration.ofHours(1),
            createdAt = Instant.parse("2024-01-15T10:00:00Z"),
            domainEventRegistry = DomainEventRegistry.delegate(),
        )
    }

    private fun createAppointmentEvent(id: AgendaEventId): AgendaEvent {
        return AgendaEvent(
            id = id,
            timeSpan =
                TimeSpan(
                    Instant.parse("2024-01-20T14:00:00Z"),
                    Instant.parse("2024-01-20T15:00:00Z"),
                ),
            attendee = Attendee("customer-456", "Jane Smith"),
            cancelReason = null,
            data =
                AppointmentEventData(
                    services = listOf(AppointmentService("Haircut"), AppointmentService("Color"))
                ),
            reminder = Reminder(id, ReminderStatus.PENDING, null),
            remindBefore = Duration.ofMinutes(30),
            createdAt = Instant.parse("2024-01-15T10:00:00Z"),
            domainEventRegistry = DomainEventRegistry.delegate(),
        )
    }
}
