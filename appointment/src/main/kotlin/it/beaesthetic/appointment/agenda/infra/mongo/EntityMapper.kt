package it.beaesthetic.appointment.agenda.infra.mongo

import it.beaesthetic.appointment.agenda.domain.event.*
import it.beaesthetic.appointment.agenda.domain.reminder.Reminder
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderOptions
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import it.beaesthetic.appointment.common.DomainEventRegistry
import java.time.Duration
import java.time.Instant

object EntityMapper {
    fun toEntity(scheduleAgenda: AgendaEvent, version: Long): AgendaEntity {
        return AgendaEntity(
            id = scheduleAgenda.id.value,
            attendee =
                AttendeeEntity(
                    id = scheduleAgenda.attendee.id,
                    displayName = scheduleAgenda.attendee.displayName
                ),
            start = scheduleAgenda.timeSpan.start,
            end = scheduleAgenda.timeSpan.end,
            cancelReason = scheduleAgenda.cancelReason,
            createdAt = scheduleAgenda.createdAt,
            updatedAt = Instant.now(),
            data =
                when (scheduleAgenda.data) {
                    is AppointmentEventData ->
                        AgendaAppointmentData(
                            services = scheduleAgenda.data.services.map { it.name }
                        )
                    is BasicEventData ->
                        AgendaBasicData(
                            title = scheduleAgenda.data.title,
                            description = scheduleAgenda.data.description ?: ""
                        )
                },
            version = version,
            reminderStatus = scheduleAgenda.activeReminder.status.name,
            isCancelled = scheduleAgenda.cancelReason != null,
            remindBeforeSeconds = scheduleAgenda.reminderOptions.triggerBefore.toSeconds().toInt()
        )
    }

    fun toDomain(scheduleAgendaEntity: AgendaEntity): AgendaEvent {
        val reminderOptions =
            ReminderOptions(
                wantRecap = false,
                triggerBefore =
                    Duration.ofSeconds(scheduleAgendaEntity.remindBeforeSeconds.toLong()),
            )
        return AgendaEvent(
            id = AgendaEventId(scheduleAgendaEntity.id),
            timeSpan = TimeSpan(scheduleAgendaEntity.start, scheduleAgendaEntity.end),
            attendee =
                Attendee(
                    id = scheduleAgendaEntity.attendee.id,
                    displayName = scheduleAgendaEntity.attendee.displayName,
                ),
            cancelReason = scheduleAgendaEntity.cancelReason,
            data =
                when (scheduleAgendaEntity.data) {
                    is AgendaAppointmentData ->
                        AppointmentEventData(
                            services =
                                scheduleAgendaEntity.data.services.map { AppointmentService(it) }
                        )
                    is AgendaBasicData ->
                        BasicEventData(
                            title = scheduleAgendaEntity.data.title,
                            description = scheduleAgendaEntity.data.description,
                        )
                },
            createdAt = scheduleAgendaEntity.createdAt,
            reminderOptions = reminderOptions,
            activeReminder =
                Reminder.from(scheduleAgendaEntity.start, reminderOptions)
                    .copy(status = ReminderStatus.valueOf(scheduleAgendaEntity.reminderStatus)),
            domainEventRegistry = DomainEventRegistry.delegate()
        )
    }
}
