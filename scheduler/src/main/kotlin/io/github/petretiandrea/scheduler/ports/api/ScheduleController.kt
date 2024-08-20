package io.github.petretiandrea.scheduler.ports.api

import io.github.petretiandrea.scheduler.api.SchedulesApi
import io.github.petretiandrea.scheduler.core.ScheduleId
import io.github.petretiandrea.scheduler.core.ScheduleMeta
import io.github.petretiandrea.scheduler.core.Scheduler
import io.github.petretiandrea.scheduler.model.AddSchedule202ResponseDto
import io.github.petretiandrea.scheduler.model.CreateScheduleDto
import java.util.*
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.RestController

@RestController
class ScheduleController(private val scheduler: Scheduler) : SchedulesApi {

    override suspend fun addSchedule(
        scheduleId: UUID,
        createScheduleDto: CreateScheduleDto
    ): ResponseEntity<AddSchedule202ResponseDto> {
        return scheduler
            .schedule(
                scheduleId.toString(),
                createScheduleDto.scheduleAt.toInstant(),
                ScheduleMeta(createScheduleDto.route, createScheduleDto.data)
            )
            .map { AddSchedule202ResponseDto(UUID.fromString(it.id.id)) }
            .map { ResponseEntity.accepted().body(it) }
            .getOrThrow()
    }

    override suspend fun removeSchedule(scheduleId: UUID): ResponseEntity<Unit> {
        return scheduler
            .delete(ScheduleId(scheduleId.toString()))
            .map { ResponseEntity.noContent().build<Unit>() }
            .getOrThrow()
    }
}
