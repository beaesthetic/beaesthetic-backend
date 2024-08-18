package it.beaesthetic.appointment.agenda.port.rest

import it.beaesthetic.appointment.agenda.domain.*
import it.beaesthetic.appointment.agenda.generated.api.model.*
import java.time.ZoneOffset
import java.util.*

object AgendaScheduleMapper {

    // TODO: fix and avoid return Any
    fun toResourceDto(schedule: AgendaSchedule): Any =
        when (schedule.data) {
            is AppointmentScheduleData -> mapAppointmentScheduleResource(schedule, schedule.data)
            is BasicScheduleData -> mapBasicScheduleResource(schedule, schedule.data)
        }

    private fun mapAppointmentScheduleResource(
        schedule: AgendaSchedule,
        scheduleData: AppointmentScheduleData
    ): AppointmentEventResponseDto {
        return AppointmentEventResponseDto(
            type = AppointmentEventResponseDto.Type.APPOINTMENT,
            id = UUID.fromString(schedule.id),
            attendee = mapAttendeeResource(schedule.attendee),
            start = schedule.timeSpan.start.atOffset(ZoneOffset.UTC),
            end = schedule.timeSpan.end.atOffset(ZoneOffset.UTC),
            isCanceled = schedule.cancelReason != null,
            reminderSent = false,
            reminder =
                AppointmentEventResponseReminderDto(
                    // TODO: implement this
                    ),
            appointment =
                AppointmentEventAppointmentDto(services = scheduleData.services.map { it.name })
        )
    }

    private fun mapBasicScheduleResource(
        schedule: AgendaSchedule,
        scheduleData: BasicScheduleData
    ): EventResponseDto {
        return EventResponseDto(
            type = EventResponseDto.Type.EVENT,
            id = UUID.fromString(schedule.id),
            attendee = mapAttendeeResource(schedule.attendee),
            start = schedule.timeSpan.start.atOffset(ZoneOffset.UTC),
            end = schedule.timeSpan.end.atOffset(ZoneOffset.UTC),
            title = scheduleData.title,
            description = scheduleData.description,
            isCanceled = schedule.cancelReason != null,
            reminderSent = false,
            reminder = AppointmentEventResponseReminderDto(), // TODO: handle this
            cancelReason = schedule.cancelReason?.let { mapCancelReasonResource(it) }
        )
    }

    private fun mapAttendeeResource(attendee: Attendee): AttendeeDto =
        AttendeeDto(
            id = UUID.fromString(attendee.id),
            name =
                attendee.displayName.split(" ").let {
                    if (it.isNotEmpty()) it[0] else attendee.displayName
                },
            surname = attendee.displayName.split(" ").let { if (it.size > 1) it[1] else "" },
        )

    private fun mapCancelReasonResource(cancelReason: CancelReason) =
        when (cancelReason) {
            CustomerCancel -> "customer_cancel"
            NoReason -> "deleted"
        }
}
