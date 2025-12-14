package it.beaesthetic.appointment.agenda.application.events

import it.beaesthetic.appointment.agenda.domain.event.*
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import java.time.Duration
import java.time.Instant
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.*

class CreateAgendaScheduleHandlerTest {

    private lateinit var agendaRepository: AgendaRepository
    private lateinit var customerRegistry: CustomerRegistry
    private lateinit var handler: CreateAgendaScheduleHandler

    @BeforeEach
    fun setup() {
        agendaRepository = mock()
        customerRegistry = mock()
        handler = CreateAgendaScheduleHandler(agendaRepository, customerRegistry)
    }

    @Test
    fun `should create agenda event with basic data`(): Unit = runBlocking {
        // Given
        val command =
            CreateAgendaSchedule(
                timeSpan =
                    TimeSpan(
                        Instant.parse("2024-01-20T10:00:00Z"),
                        Instant.parse("2024-01-20T11:00:00Z"),
                    ),
                data = BasicEventData("Meeting", "Important meeting"),
                attendeeId = "user-123",
                triggerBefore = Duration.ofHours(1),
            )

        whenever(agendaRepository.saveEvent(any(), any())).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isSuccess)
        val event = result.getOrThrow()

        assertEquals(command.timeSpan, event.timeSpan)
        assertEquals(command.data, event.data)
        assertEquals(Attendee("user-123", "self"), event.attendee)
        assertEquals(command.triggerBefore, event.remindBefore)
        assertEquals(ReminderStatus.PENDING, event.reminder.status)

        verify(agendaRepository).saveEvent(any(), eq(0L))
    }

    @Test
    fun `should create agenda event with appointment data and fetch customer`(): Unit =
        runBlocking {
            // Given
            val customerId = "customer-456"
            val customer = Customer(customerId, "John Doe", "+1234567890")
            val services = listOf(AppointmentService("Haircut"), AppointmentService("Massage"))

            val command =
                CreateAgendaSchedule(
                    timeSpan =
                        TimeSpan(
                            Instant.parse("2024-01-20T14:00:00Z"),
                            Instant.parse("2024-01-20T15:30:00Z"),
                        ),
                    data = AppointmentEventData(services),
                    attendeeId = customerId,
                    triggerBefore = Duration.ofMinutes(30),
                )

            whenever(customerRegistry.findByCustomerId(customerId)).thenReturn(customer)

            whenever(agendaRepository.saveEvent(any(), any())).thenAnswer { invocation ->
                invocation.getArgument<AgendaEvent>(0)
            }

            // When
            val result = handler.handle(command)

            // Then
            assertTrue(result.isSuccess)
            val event = result.getOrThrow()

            assertEquals(Attendee(customerId, "John Doe"), event.attendee)
            assertEquals(command.data, event.data)
            assertEquals("Haircut,Massage", event.title)

            verify(customerRegistry).findByCustomerId(customerId)
            verify(agendaRepository).saveEvent(any(), eq(0L))
        }

    @Test
    fun `should fail when customer not found for appointment data`(): Unit = runBlocking {
        // Given
        val customerId = "non-existent-customer"
        val command =
            CreateAgendaSchedule(
                timeSpan =
                    TimeSpan(
                        Instant.parse("2024-01-20T14:00:00Z"),
                        Instant.parse("2024-01-20T15:00:00Z"),
                    ),
                data = AppointmentEventData(listOf(AppointmentService("Haircut"))),
                attendeeId = customerId,
                triggerBefore = Duration.ofHours(1),
            )

        whenever(customerRegistry.findByCustomerId(customerId)).thenReturn(null)

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull() is IllegalArgumentException)
        assertTrue(result.exceptionOrNull()!!.message!!.contains("Unknown attendee"))

        verify(customerRegistry).findByCustomerId(customerId)
        verify(agendaRepository, never()).saveEvent(any(), any())
    }

    @Test
    fun `should fail when repository save fails`(): Unit = runBlocking {
        // Given
        val command =
            CreateAgendaSchedule(
                timeSpan =
                    TimeSpan(
                        Instant.parse("2024-01-20T10:00:00Z"),
                        Instant.parse("2024-01-20T11:00:00Z"),
                    ),
                data = BasicEventData("Meeting", "Test"),
                attendeeId = "user-123",
                triggerBefore = Duration.ofHours(1),
            )

        val repositoryError = RuntimeException("Database connection failed")
        whenever(agendaRepository.saveEvent(any(), any()))
            .thenReturn(Result.failure(repositoryError))

        // When
        val result = handler.handle(command)

        // Then
        assertTrue(result.isFailure)
        assertEquals(repositoryError, result.exceptionOrNull())

        verify(agendaRepository).saveEvent(any(), eq(0L))
    }

    @Test
    fun `should generate unique event ID for each creation`() = runBlocking {
        // Given
        val command =
            CreateAgendaSchedule(
                timeSpan =
                    TimeSpan(
                        Instant.parse("2024-01-20T10:00:00Z"),
                        Instant.parse("2024-01-20T11:00:00Z"),
                    ),
                data = BasicEventData("Meeting", "Test"),
                attendeeId = "user-123",
                triggerBefore = Duration.ofHours(1),
            )

        whenever(agendaRepository.saveEvent(any(), any())).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result1 = handler.handle(command)
        val result2 = handler.handle(command)

        // Then
        assertTrue(result1.isSuccess)
        assertTrue(result2.isSuccess)

        val event1 = result1.getOrThrow()
        val event2 = result2.getOrThrow()

        // Event IDs should be different (UUIDs)
        assertTrue(event1.id.value != event2.id.value)
    }

    @Test
    fun `should set reminder to PENDING status on creation`() = runBlocking {
        // Given
        val command =
            CreateAgendaSchedule(
                timeSpan =
                    TimeSpan(
                        Instant.parse("2024-01-20T10:00:00Z"),
                        Instant.parse("2024-01-20T11:00:00Z"),
                    ),
                data = BasicEventData("Test", null),
                attendeeId = "user-123",
                triggerBefore = Duration.ofMinutes(15),
            )

        whenever(agendaRepository.saveEvent(any(), any())).thenAnswer { invocation ->
            invocation.getArgument<AgendaEvent>(0)
        }

        // When
        val result = handler.handle(command)

        // Then
        val event = result.getOrThrow()
        assertEquals(ReminderStatus.PENDING, event.reminder.status)
        assertEquals(event.id, event.reminder.eventId)
        assertEquals(null, event.reminder.sentAt)
    }
}
