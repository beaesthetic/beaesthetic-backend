package it.beaesthetic.appointment.agenda.port.rest

import io.smallrye.mutiny.Uni
import it.beaesthetic.appointment.agenda.application.events.*
import it.beaesthetic.appointment.agenda.domain.event.*
import it.beaesthetic.appointment.agenda.generated.api.ActivitiesApi
import it.beaesthetic.appointment.agenda.generated.api.model.*
import it.beaesthetic.appointment.service.common.uniWithScope
import java.time.OffsetDateTime
import java.util.*
import kotlin.IllegalArgumentException as IllegalArgumentException1

class AgendaController(
    private val createAgendaScheduleHandler: CreateAgendaScheduleHandler,
    private val editAgendaScheduleHandler: EditAgendaScheduleHandler,
    private val deleteAgendaScheduleHandler: DeleteAgendaScheduleHandler,
    private val queryHandler: AgendaQueryHandler
) : ActivitiesApi {

    override fun createAgendaActivity(
        createAgendaActivityMixin: CreateAgendaActivityMixin
    ): Uni<CreateAgendaActivity201ResponseDto> = uniWithScope {
        val command =
            when (createAgendaActivityMixin) {
                is GenericEventDto ->
                    CreateAgendaSchedule(
                        timeSpan =
                            TimeSpan.fromOffsetDateTime(
                                createAgendaActivityMixin.start,
                                createAgendaActivityMixin.end
                            ),
                        attendeeId = createAgendaActivityMixin.attendeeId.toString(),
                        data =
                            BasicEventData(
                                title = createAgendaActivityMixin.title,
                                description = createAgendaActivityMixin.description
                            )
                    )
                is AppointmentEventDto ->
                    CreateAgendaSchedule(
                        timeSpan =
                            TimeSpan.fromOffsetDateTime(
                                createAgendaActivityMixin.start,
                                createAgendaActivityMixin.end
                            ),
                        attendeeId = createAgendaActivityMixin.attendeeId.toString(),
                        data =
                            AppointmentEventData(
                                services =
                                    createAgendaActivityMixin.appointment.services?.map {
                                        AppointmentService(it)
                                    }
                                        ?: emptyList()
                            )
                    )
                else ->
                    throw IllegalArgumentException1(
                        "Unrecognized request type: ${createAgendaActivityMixin.javaClass.name}"
                    )
            }

        createAgendaScheduleHandler
            .handle(command)
            .map { CreateAgendaActivity201ResponseDto(UUID.fromString(it.id)) }
            .getOrThrow()
    }

    override fun updateActivity(
        activityId: UUID,
        updateActivityRequestDto: UpdateActivityRequestDto
    ): Uni<ActivityResponseDto> = uniWithScope {
        val command =
            EditAgendaSchedule(
                scheduleId = activityId.toString(),
                description = updateActivityRequestDto.description,
                title = updateActivityRequestDto.title,
                services =
                    updateActivityRequestDto.services?.map { AppointmentService(it) }?.toSet(),
                timeSpan =
                    updateActivityRequestDto.start?.let { start ->
                        updateActivityRequestDto.end?.let { end ->
                            TimeSpan.fromOffsetDateTime(start, end)
                        }
                    }
            )

        editAgendaScheduleHandler
            .handle(command)
            .map { AgendaScheduleMapper.toResourceDto(it) }
            .getOrThrow()
    }

    override fun getActivityById(activityId: UUID): Uni<ActivityResponseDto> = uniWithScope {
        queryHandler
            .handle(activityId.toString())
            .map { AgendaScheduleMapper.toResourceDto(it) }
            .getOrThrow()
    }

    override fun deleteActivity(activityId: UUID, reason: String?): Uni<Void> =
        uniWithScope {
                deleteAgendaScheduleHandler
                    .handle(
                        DeleteAgendaSchedule(
                            activityId,
                            when (reason) {
                                "customer_cancel" -> CustomerCancel
                                "deleted" -> NoReason
                                else -> NoReason
                            }
                        )
                    )
                    .getOrThrow()
            }
            .replaceWithVoid()

    override fun getAppointmentsByTimeRangeOrCustomer(
        start: OffsetDateTime?,
        end: OffsetDateTime?,
        attendeeId: String?
    ): Uni<ActivityResponseDto> = uniWithScope {
        val timeSpan =
            start?.let {
                end?.let { TimeSpan.fromOffsetDateTime(start, end) }
                    ?: throw IllegalArgumentException("end must be specified")
            }

        when {
            timeSpan != null -> queryHandler.handle(timeSpan)
            attendeeId != null -> queryHandler.handleByAttendeeId(attendeeId)
            else -> emptyList()
        }.map { AgendaScheduleMapper.toResourceDto(it) }
    }
}
