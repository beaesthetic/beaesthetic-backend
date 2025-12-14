package it.beaesthetic.appointment.agenda.infra.mongo

import it.beaesthetic.appointment.agenda.domain.event.*
import it.beaesthetic.appointment.agenda.domain.reminder.Reminder
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
                    displayName = scheduleAgenda.attendee.displayName,
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
                            description = scheduleAgenda.data.description ?: "",
                        )
                },
            version = version,
            reminderStatus = scheduleAgenda.reminder.status.name,
            remindBeforeSeconds = scheduleAgenda.remindBefore.toSeconds().toInt(),
            reminderSentAt = scheduleAgenda.reminder.sentAt,
            isCancelled = scheduleAgenda.cancelReason != null,
        )
    }

    fun toDomain(scheduleAgendaEntity: AgendaEntity): AgendaEvent {
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
            reminder =
                Reminder(
                    eventId = AgendaEventId(scheduleAgendaEntity.id),
                    status = ReminderStatus.valueOf(scheduleAgendaEntity.reminderStatus),
                    sentAt = scheduleAgendaEntity.reminderSentAt,
                ),
            remindBefore = Duration.ofSeconds(scheduleAgendaEntity.remindBeforeSeconds.toLong()),
            domainEventRegistry = DomainEventRegistry.delegate(),
        )
    }
}
