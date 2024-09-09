package it.beaesthetic.appointment.agenda.port.rest

import it.beaesthetic.appointment.agenda.domain.event.*
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import it.beaesthetic.appointment.agenda.generated.api.model.*
import java.time.ZoneOffset
import java.util.*

object AgendaScheduleMapper {

    // TODO: fix and avoid return Any
    fun toResourceDto(schedule: AgendaEvent): Any =
        when (schedule.data) {
            is AppointmentEventData -> mapAppointmentScheduleResource(schedule, schedule.data)
            is BasicEventData -> mapBasicScheduleResource(schedule, schedule.data)
        }

    private fun mapAppointmentScheduleResource(
        schedule: AgendaEvent,
        scheduleData: AppointmentEventData
    ): AppointmentEventResponseDto {
        return AppointmentEventResponseDto(
            type = AppointmentEventResponseDto.Type.APPOINTMENT,
            id = UUID.fromString(schedule.id.value),
            attendee = mapAttendeeResource(schedule.attendee),
            start = schedule.timeSpan.start.atOffset(ZoneOffset.UTC),
            end = schedule.timeSpan.end.atOffset(ZoneOffset.UTC),
            isCanceled = schedule.cancelReason != null,
            reminderSent = schedule.activeReminder.status == ReminderStatus.SENT,
            reminder =
                AppointmentEventResponseReminderDto(
                    status = mapReminderStatusResource(schedule.activeReminder.status),
                    reminderMinutes = schedule.reminderOptions.triggerBefore.toMinutes().toInt(),
                    timeSent = schedule.activeReminder.timeToSend.atOffset(ZoneOffset.UTC)
                ),
            appointment =
                AppointmentEventAppointmentDto(services = scheduleData.services.map { it.name }),
            cancelReason = schedule.cancelReason?.let { mapCancelReasonResource(it) }
        )
    }

    private fun mapBasicScheduleResource(
        schedule: AgendaEvent,
        scheduleData: BasicEventData
    ): EventResponseDto {
        return EventResponseDto(
            type = EventResponseDto.Type.EVENT,
            id = UUID.fromString(schedule.id.value),
            attendee = mapAttendeeResource(schedule.attendee),
            start = schedule.timeSpan.start.atOffset(ZoneOffset.UTC),
            end = schedule.timeSpan.end.atOffset(ZoneOffset.UTC),
            title = scheduleData.title,
            description = scheduleData.description,
            isCanceled = schedule.cancelReason != null,
            reminderSent = schedule.activeReminder.status == ReminderStatus.SENT,
            reminder =
                AppointmentEventResponseReminderDto(
                    status = mapReminderStatusResource(schedule.activeReminder.status),
                    reminderMinutes = schedule.reminderOptions.triggerBefore.toMinutes().toInt(),
                    timeSent = schedule.activeReminder.timeToSend.atOffset(ZoneOffset.UTC)
                ),
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

    private fun mapReminderStatusResource(reminderStatus: ReminderStatus): AppointmentEventResponseReminderDto.Status {
        return when(reminderStatus) {
            ReminderStatus.SENT -> AppointmentEventResponseReminderDto.Status.SENT
            ReminderStatus.SENT_REQUESTED -> AppointmentEventResponseReminderDto.Status.SEND_IN_PROGRESS
            ReminderStatus.DELETED,
            ReminderStatus.SCHEDULED,
            ReminderStatus.PENDING -> AppointmentEventResponseReminderDto.Status.NOT_SENT
        }
    }
}
