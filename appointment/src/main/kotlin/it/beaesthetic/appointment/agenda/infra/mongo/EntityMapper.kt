package it.beaesthetic.appointment.agenda.infra.mongo

import it.beaesthetic.appointment.agenda.domain.*
import java.time.Instant

object EntityMapper {
    fun toEntity(scheduleAgenda: AgendaSchedule, version: Long): AgendaEntity {
        return AgendaEntity(
            id = scheduleAgenda.id,
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
                    is AppointmentScheduleData ->
                        AgendaAppointmentData(
                            services = scheduleAgenda.data.services.map { it.name }
                        )
                    is BasicScheduleData ->
                        AgendaBasicData(
                            title = scheduleAgenda.data.title,
                            description = scheduleAgenda.data.description ?: ""
                        )
                },
            version = version
        )
    }

    fun toDomain(scheduleAgendaEntity: AgendaEntity): AgendaSchedule {
        return AgendaSchedule(
            id = scheduleAgendaEntity.id,
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
                        AppointmentScheduleData(
                            services =
                                scheduleAgendaEntity.data.services.map { AppointmentService(it) }
                        )
                    is AgendaBasicData ->
                        BasicScheduleData(
                            title = scheduleAgendaEntity.data.title,
                            description = scheduleAgendaEntity.data.description,
                        )
                },
            createdAt = scheduleAgendaEntity.createdAt
        )
    }
}
