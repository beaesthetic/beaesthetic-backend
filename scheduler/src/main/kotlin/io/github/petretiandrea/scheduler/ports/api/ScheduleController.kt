package io.github.petretiandrea.scheduler.ports.api

import io.github.petretiandrea.scheduler.api.SchedulesApi
import io.github.petretiandrea.scheduler.core.ScheduleId
import io.github.petretiandrea.scheduler.core.ScheduleMeta
import io.github.petretiandrea.scheduler.core.Scheduler
import io.github.petretiandrea.scheduler.model.AddSchedule202ResponseDto
import io.github.petretiandrea.scheduler.model.CreateScheduleDto
import java.util.*
import org.slf4j.LoggerFactory
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.RestController

@RestController
class ScheduleController(private val scheduler: Scheduler) : SchedulesApi {

    private val log = LoggerFactory.getLogger(ScheduleController::class.java)

    override suspend fun addSchedule(
        scheduleId: UUID,
        createScheduleDto: CreateScheduleDto,
    ): ResponseEntity<AddSchedule202ResponseDto> {
        log.info(
            "Adding schedule {} at {} will delivered on {}",
            scheduleId,
            createScheduleDto.scheduleAt,
            createScheduleDto.route,
        )
        return scheduler
            .schedule(
                scheduleId.toString(),
                createScheduleDto.scheduleAt.toInstant(),
                ScheduleMeta(createScheduleDto.route, createScheduleDto.data),
            )
            .map { AddSchedule202ResponseDto(UUID.fromString(it.id.id)) }
            .map { ResponseEntity.accepted().body(it) }
            .getOrThrow()
    }

    override suspend fun removeSchedule(scheduleId: UUID): ResponseEntity<Unit> {
        log.info("Removing schedule {}", scheduleId)
        return scheduler
            .delete(ScheduleId(scheduleId.toString()))
            .map { ResponseEntity.noContent().build<Unit>() }
            .getOrThrow()
    }
}
