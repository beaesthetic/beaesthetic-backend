package it.beaesthetic.appointment.agenda.port.rest

import io.smallrye.mutiny.Uni
import it.beaesthetic.appointment.agenda.application.CreateAgendaSchedule
import it.beaesthetic.appointment.agenda.application.CreateAgendaScheduleHandler
import it.beaesthetic.appointment.agenda.application.EditAgendaSchedule
import it.beaesthetic.appointment.agenda.application.EditAgendaScheduleHandler
import it.beaesthetic.appointment.agenda.domain.*
import it.beaesthetic.appointment.agenda.generated.api.ActivitiesApi
import it.beaesthetic.appointment.agenda.generated.api.model.*
import it.beaesthetic.appointment.service.common.uniWithScope
import java.time.OffsetDateTime
import java.util.*
import kotlin.IllegalArgumentException as IllegalArgumentException1

class AgendaController(
    private val createAgendaScheduleHandler: CreateAgendaScheduleHandler,
    private val editAgendaScheduleHandler: EditAgendaScheduleHandler
) : ActivitiesApi {

    override fun createAgendaActivity(
        createAgendaActivityMixin: CreateAgendaActivityMixin
    ): Uni<CreateAgendaActivity201ResponseDto> = uniWithScope {
        val command = when (createAgendaActivityMixin) {
            is GenericEventDto -> CreateAgendaSchedule(
                timeSpan = TimeSpan.fromOffsetDateTime(createAgendaActivityMixin.start, createAgendaActivityMixin.end),
                attendeeId = createAgendaActivityMixin.attendeeId.toString(),
                data = BasicScheduleData(
                    title = createAgendaActivityMixin.title,
                    description = createAgendaActivityMixin.description
                )
            )
            is AppointmentEventDto -> CreateAgendaSchedule(
                timeSpan = TimeSpan.fromOffsetDateTime(createAgendaActivityMixin.start, createAgendaActivityMixin.end),
                attendeeId = createAgendaActivityMixin.attendeeId.toString(),
                data = AppointmentScheduleData(
                    services = createAgendaActivityMixin.appointment.services
                        ?.map { AppointmentService(it) } ?: emptyList()
                )
            )
            else -> throw IllegalArgumentException1("Unrecognized request type: ${createAgendaActivityMixin.javaClass.name}")
        }

        createAgendaScheduleHandler.handle(command)
            .map { CreateAgendaActivity201ResponseDto(UUID.fromString(it.id)) }
            .getOrThrow()
    }


    override fun deleteActivity(activityId: UUID, reason: String?): Uni<Void> {
        TODO("Not yet implemented")
    }

    override fun getAppointmentsByTimeRangeOrCustomer(
        start: OffsetDateTime?,
        end: OffsetDateTime?,
        attendeeId: String?
    ): Uni<ActivityResponseDto> {
        TODO("Not yet implemented")
    }

    override fun updateActivity(
        activityId: UUID,
        updateActivityRequestDto: UpdateActivityRequestDto
    ): Uni<ActivityResponseDto> = uniWithScope {
        val command = EditAgendaSchedule(
            scheduleId = activityId.toString(),
            description = updateActivityRequestDto.description,
            title = updateActivityRequestDto.title,
            services = updateActivityRequestDto.services?.map { AppointmentService(it) }?.toSet(),
            timeSpan = updateActivityRequestDto.start?.let { start ->
                updateActivityRequestDto.end?.let { end ->
                    TimeSpan.fromOffsetDateTime(start, end)
                }
            }
        )

        editAgendaScheduleHandler.handle(command)
            .map {  }
            .getOrThrow()
    }


}