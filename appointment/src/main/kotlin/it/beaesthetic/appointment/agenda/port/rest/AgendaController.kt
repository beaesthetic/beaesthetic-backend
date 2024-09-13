package it.beaesthetic.appointment.agenda.port.rest

import io.quarkus.runtime.annotations.RegisterForReflection
import io.smallrye.mutiny.Uni
import it.beaesthetic.appointment.agenda.application.events.*
import it.beaesthetic.appointment.agenda.config.ReminderConfiguration
import it.beaesthetic.appointment.agenda.domain.event.*
import it.beaesthetic.appointment.agenda.generated.api.ActivitiesApi
import it.beaesthetic.appointment.agenda.generated.api.model.*
import it.beaesthetic.appointment.service.common.uniWithScope
import java.time.OffsetDateTime
import java.util.*
import kotlin.IllegalArgumentException as IllegalArgumentException1

@RegisterForReflection(
    targets =
        [
            CreateAgendaActivity201ResponseDto::class,
            ActivityResponseDto::class,
        ],
    registerFullHierarchy = true
)
class AgendaController(
    private val createAgendaScheduleHandler: CreateAgendaScheduleHandler,
    private val editAgendaScheduleHandler: EditAgendaScheduleHandler,
    private val deleteAgendaScheduleHandler: DeleteAgendaScheduleHandler,
    private val queryHandler: AgendaQueryHandler,
    private val reminderConfiguration: ReminderConfiguration
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
                        triggerBefore = reminderConfiguration.triggerBefore(),
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
                        triggerBefore = reminderConfiguration.triggerBefore(),
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
            .map { CreateAgendaActivity201ResponseDto(UUID.fromString(it.id.value)) }
            .getOrThrow()
    }

    override fun updateActivity(
        activityId: UUID,
        updateActivityRequestDto: UpdateActivityRequestDto
    ): Uni<ActivityResponseDto> = uniWithScope {
        val command =
            EditAgendaSchedule(
                scheduleId = AgendaEventId(activityId.toString()),
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
            .handle(AgendaEventId(activityId.toString()))
            .map { AgendaScheduleMapper.toResourceDto(it) }
            .getOrThrow()
    }

    override fun deleteActivity(activityId: UUID, reason: String?): Uni<Void> =
        uniWithScope {
                deleteAgendaScheduleHandler
                    .handle(
                        DeleteAgendaSchedule(
                            AgendaEventId(activityId.toString()),
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
