package it.beaesthetic.appointment.agenda.domain

import java.time.Instant

data class Attendee(val id: String, val displayName: String)

data class AgendaSchedule(
    val id: String,
    val timeSpan: TimeSpan,
    val attendee: Attendee,
    val cancelReason: CancelReason?,
    val data: AgendaScheduleData,
    val createdAt: Instant,
) {

    val title = when(data) {
        is BasicScheduleData -> data.title
        is AppointmentScheduleData -> data.services.joinToString(",") { it.name }
    }

    fun reschedule(timeSpan: TimeSpan): AgendaSchedule {
        require(cancelReason == null) { "Event already canceled"}
        return copy(timeSpan = timeSpan)
    }

    fun cancel(reason: CancelReason): AgendaSchedule {
        require(cancelReason == null) { "Event already canceled"}
        return copy(
            cancelReason = reason
        )
    }
}


