package it.beaesthetic.appointment.agenda.infra

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.convertValue
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.domain.reminder.Reminder
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderScheduler
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderTimesUp
import it.beaesthetic.generated.scheduler.client.api.SchedulesApi
import it.beaesthetic.generated.scheduler.client.model.CreateSchedule
import java.time.Instant
import java.time.ZoneOffset
import java.util.*

class RemoteScheduler(
    private val schedulesApi: SchedulesApi,
    private val schedulerRoute: String,
    private val objectMapper: ObjectMapper
) : ReminderScheduler {

    override suspend fun scheduleReminder(reminder: Reminder, sendAt: Instant): Reminder {
        val dataMap: Map<String, Any> =
            objectMapper.convertValue(ReminderTimesUp(reminder.eventId.value))
        val request =
            CreateSchedule().apply {
                scheduleAt = sendAt.atOffset(ZoneOffset.UTC)
                route = schedulerRoute
                data = dataMap
            }
        return schedulesApi
            .addSchedule(UUID.fromString(reminder.eventId.value), request)
            .awaitSuspending()
            .let { reminder }
    }

    override suspend fun unschedule(reminder: Reminder): Reminder {
        TODO("Not yet implemented")
    }
}
