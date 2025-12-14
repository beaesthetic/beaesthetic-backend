package it.beaesthetic.appointment.agenda.application.events

import it.beaesthetic.appointment.agenda.domain.event.*
import it.beaesthetic.appointment.agenda.domain.reminder.Reminder
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import it.beaesthetic.appointment.common.DomainEventRegistry
import it.beaesthetic.appointment.common.OptimisticConcurrency.VersionedEntity
import java.time.Duration
import java.time.Instant
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.*

class DeleteAgendaScheduleHandlerTest {

    private lateinit var agendaRepository: AgendaRepository
    private lateinit var handler: DeleteAgendaScheduleHandler

    @BeforeEach
    fun setup() {
        agendaRepository = mock()
        handler = DeleteAgendaScheduleHandler(agendaRepository)
    }

    @Test
    fun `should cancel event with customer cancel reason`(): Unit = runBlocking {
        // Given
        val eventId = AgendaEventId("event-123")
        val existingEvent = createTestEvent(eventId)
        val version = 5L

        val command = DeleteAgendaSchedule(id = eventId, reason = CustomerCancel)

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        whenever(agendaRepository.saveEvent(any(), eq(version))).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isSuccess)
        val canceledEvent = result.getOrThrow()

        assertEquals(CustomerCancel, canceledEvent.cancelReason)
        assertEquals(eventId, canceledEvent.id)
        assertEquals(existingEvent.timeSpan, canceledEvent.timeSpan)
        assertEquals(existingEvent.attendee, canceledEvent.attendee)

        verify(agendaRepository).saveEvent(any(), eq(version))
    }

    @Test
    fun `should cancel event with no reason`(): Unit = runBlocking {
        // Given
        val eventId = AgendaEventId("event-456")
        val existingEvent = createTestEvent(eventId)
        val version = 3L

        val command = DeleteAgendaSchedule(id = eventId, reason = NoReason)

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        whenever(agendaRepository.saveEvent(any(), eq(version))).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isSuccess)
        val canceledEvent = result.getOrThrow()

        assertEquals(NoReason, canceledEvent.cancelReason)
        verify(agendaRepository).saveEvent(any(), eq(version))
    }

    @Test
    fun `should throw exception when event not found`(): Unit = runBlocking {
        // Given
        val eventId = AgendaEventId("non-existent")
        val command = DeleteAgendaSchedule(id = eventId, reason = CustomerCancel)

        whenever(agendaRepository.findEvent(eventId)).thenReturn(null)

        // When & Then
        val exception = runCatching { handler.handle(command) }.exceptionOrNull()

        assertNotNull(exception)
        assertTrue(exception is IllegalArgumentException)
        assertTrue(exception.message!!.contains("Schedule not found"))

        verify(agendaRepository, never()).saveEvent(any(), any())
    }

    @Test
    fun `should use optimistic locking with correct version`(): Unit = runBlocking {
        // Given
        val eventId = AgendaEventId("event-123")
        val existingEvent = createTestEvent(eventId)
        val version = 99L

        val command = DeleteAgendaSchedule(id = eventId, reason = CustomerCancel)

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        whenever(agendaRepository.saveEvent(any(), eq(version))).thenAnswer { invocation ->
            Result.success(invocation.getArgument<AgendaEvent>(0))
        }

        // When
        handler.handle(command)

        // Then
        verify(agendaRepository).saveEvent(any(), eq(99L))
    }

    @Test
    fun `should fail when repository save fails`() = runBlocking {
        // Given
        val eventId = AgendaEventId("event-123")
        val existingEvent = createTestEvent(eventId)
        val version = 2L

        val command = DeleteAgendaSchedule(id = eventId, reason = CustomerCancel)

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        val repositoryError = RuntimeException("Database connection lost")
        whenever(agendaRepository.saveEvent(any(), eq(version)))
            .thenReturn(Result.failure(repositoryError))

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isFailure)
        assertEquals(repositoryError, result.exceptionOrNull())
    }

    @Test
    fun `should preserve event data when canceling`() = runBlocking {
        // Given
        val eventId = AgendaEventId("event-123")
        val originalData =
            AppointmentEventData(
                services = listOf(AppointmentService("Haircut"), AppointmentService("Styling"))
            )
        val existingEvent = createTestEvent(eventId, data = originalData)
        val version = 1L

        val command = DeleteAgendaSchedule(id = eventId, reason = CustomerCancel)

        whenever(agendaRepository.findEvent(eventId))
            .thenReturn(VersionedEntity(existingEvent, version))

        whenever(agendaRepository.saveEvent(any(), eq(version))).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result = handler.handle(command)

        // Then
        val canceledEvent = result.getOrThrow()

        assertEquals(originalData, canceledEvent.data)
        assertEquals(existingEvent.timeSpan, canceledEvent.timeSpan)
        assertEquals(existingEvent.attendee, canceledEvent.attendee)
        assertEquals(existingEvent.remindBefore, canceledEvent.remindBefore)
    }

    private fun createTestEvent(
        id: AgendaEventId,
        data: AgendaEventData = BasicEventData("Test Meeting", "Test Description"),
    ): AgendaEvent {
        return AgendaEvent(
            id = id,
            timeSpan =
                TimeSpan(
                    Instant.parse("2024-01-20T10:00:00Z"),
                    Instant.parse("2024-01-20T11:00:00Z"),
                ),
            attendee = Attendee("customer-123", "Test Customer"),
            cancelReason = null,
            data = data,
            reminder = Reminder(id, ReminderStatus.PENDING, null),
            remindBefore = Duration.ofHours(1),
            createdAt = Instant.parse("2024-01-15T10:00:00Z"),
            domainEventRegistry = DomainEventRegistry.delegate(),
        )
    }
}
