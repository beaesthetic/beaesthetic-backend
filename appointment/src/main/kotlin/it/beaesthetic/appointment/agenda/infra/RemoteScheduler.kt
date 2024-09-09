package it.beaesthetic.appointment.agenda.infra

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.convertValue
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderScheduler
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderTimesUp
import it.beaesthetic.generated.scheduler.client.api.SchedulesApi
import it.beaesthetic.generated.scheduler.client.model.CreateSchedule
import java.time.ZoneOffset
import java.util.*

class RemoteScheduler(
    private val schedulesApi: SchedulesApi,
    private val objectMapper: ObjectMapper
) : ReminderScheduler {

    override suspend fun scheduleReminder(event: AgendaEvent): AgendaEvent {
        val dataMap: Map<String, Any> = objectMapper.convertValue(ReminderTimesUp(event.id.value))
        val request =
            CreateSchedule().apply {
                scheduleAt = event.activeReminder.timeToSend.atOffset(ZoneOffset.UTC)
                route = "reminders"
                data = dataMap
            }
        return schedulesApi
            .addSchedule(UUID.fromString(event.id.value), request)
            .awaitSuspending()
            .let { event }
    }

    override suspend fun unschedule(event: AgendaEvent): AgendaEvent {
        TODO("Not yet implemented")
    }
}
