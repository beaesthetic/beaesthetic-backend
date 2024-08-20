package io.github.petretiandrea.scheduler.core

import java.time.Instant

@JvmInline value class ScheduleId(val id: String)

data class ScheduleMeta(val route: String, val data: Map<String, Any>)

data class ScheduleJob(val id: ScheduleId, val meta: ScheduleMeta, val scheduleAt: Instant)
