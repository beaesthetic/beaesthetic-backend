package it.beaesthetic.appointment.agenda.port.rest

import it.beaesthetic.appointment.agenda.domain.AgendaSchedule
import it.beaesthetic.appointment.agenda.domain.AppointmentScheduleData
import it.beaesthetic.appointment.agenda.domain.BasicScheduleData
import it.beaesthetic.appointment.agenda.generated.api.model.*
import java.time.ZoneOffset
import java.util.*

object AgendaScheduleMapper {

    fun toResourceDto(schedule: AgendaSchedule): AppointmentEventResponseDto = when(schedule.data) {
        is AppointmentScheduleData -> AppointmentEventResponseDto(
            type = AppointmentEventResponseDto.Type.APPOINTMENT,
            id = UUID.fromString(schedule.id),
            attendee = AttendeeDto(
                id = UUID.fromString(schedule.attendee.id),
                name = schedule.attendee.displayName.split(" ")[0],
                surname = ""
            ),
            start = schedule.timeSpan.start.atOffset(ZoneOffset.UTC),
            end = schedule.timeSpan.end.atOffset(ZoneOffset.UTC),
            isCanceled = schedule.cancelReason != null,
            reminderSent = false,
            reminder = AppointmentEventResponseReminderDto(
                // TODO: implement this
            ),
            appointment = when(schedule.data) {
                is AppointmentScheduleData -> AppointmentEventAppointmentDto(
                    services = schedule.data.services.map { it.name }
                )
                else -> AppointmentEventAppointmentDto(services = emptyList())
            }
        )
        is BasicScheduleData -> AppointmentEventResponseDto(
            type = AppointmentEventResponseDto.Type.APPOINTMENT,
            id = UUID.fromString(schedule.id),
            attendee = AttendeeDto(
                id = UUID.fromString(schedule.attendee.id),
                name = schedule.attendee.displayName.split(" ")[0],
                surname = ""
            ),
            start = schedule.timeSpan.start.atOffset(ZoneOffset.UTC),
            end = schedule.timeSpan.end.atOffset(ZoneOffset.UTC),
            isCanceled = schedule.cancelReason != null,
            reminderSent = false,
            reminder = AppointmentEventResponseReminderDto(
                // TODO: implement this
            ),
            appointment = when(schedule.data) {
                is AppointmentScheduleData -> AppointmentEventAppointmentDto(
                    services = schedule.data.services.map { it.name }
                )
                else -> AppointmentEventAppointmentDto(services = emptyList())
            }
        )
    }

}