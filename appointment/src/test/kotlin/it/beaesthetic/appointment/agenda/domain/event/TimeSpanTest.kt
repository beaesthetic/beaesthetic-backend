package it.beaesthetic.appointment.agenda.domain.event

import java.time.Instant
import java.time.OffsetDateTime
import java.time.ZoneOffset
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

class TimeSpanTest {

    @Test
    fun `should create valid TimeSpan when start is before end`() {
        // Given
        val start = Instant.parse("2024-01-15T10:00:00Z")
        val end = Instant.parse("2024-01-15T11:00:00Z")

        // When
        val timeSpan = TimeSpan(start, end)

        // Then
        assertEquals(start, timeSpan.start)
        assertEquals(end, timeSpan.end)
    }

    @Test
    fun `should create valid TimeSpan when start equals end`() {
        // Given
        val instant = Instant.parse("2024-01-15T10:00:00Z")

        // When
        val timeSpan = TimeSpan(instant, instant)

        // Then
        assertEquals(instant, timeSpan.start)
        assertEquals(instant, timeSpan.end)
    }

    @Test
    fun `should throw exception when start is after end`() {
        // Given
        val start = Instant.parse("2024-01-15T11:00:00Z")
        val end = Instant.parse("2024-01-15T10:00:00Z")

        // When & Then
        val exception = assertThrows<IllegalArgumentException> { TimeSpan(start, end) }

        assertTrue(exception.message!!.contains("must be before"))
    }

    @Test
    fun `should create TimeSpan from OffsetDateTime`() {
        // Given
        val startOdt = OffsetDateTime.of(2024, 1, 15, 10, 0, 0, 0, ZoneOffset.UTC)
        val endOdt = OffsetDateTime.of(2024, 1, 15, 11, 30, 0, 0, ZoneOffset.UTC)

        // When
        val timeSpan = TimeSpan.fromOffsetDateTime(startOdt, endOdt)

        // Then
        assertEquals(startOdt.toInstant(), timeSpan.start)
        assertEquals(endOdt.toInstant(), timeSpan.end)
    }

    @Test
    fun `should handle different timezones correctly`() {
        // Given
        val startUTC = OffsetDateTime.of(2024, 1, 15, 10, 0, 0, 0, ZoneOffset.UTC)
        val endCET = OffsetDateTime.of(2024, 1, 15, 12, 0, 0, 0, ZoneOffset.ofHours(1))

        // When
        val timeSpan = TimeSpan.fromOffsetDateTime(startUTC, endCET)

        // Then - Both should be converted to Instant correctly
        assertEquals(startUTC.toInstant(), timeSpan.start)
        assertEquals(endCET.toInstant(), timeSpan.end)
        assertTrue(timeSpan.start.isBefore(timeSpan.end))
    }
}
