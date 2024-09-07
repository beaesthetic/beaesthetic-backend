package it.beaesthetic.appointment.agenda.domain.event

import it.beaesthetic.appointment.agenda.domain.reminder.Reminder
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderOptions
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import it.beaesthetic.appointment.common.DomainEventRegistry
import java.time.Instant

data class Attendee(val id: String, val displayName: String)

data class AgendaEvent(
    val id: String,
    val timeSpan: TimeSpan,
    val attendee: Attendee,
    val cancelReason: CancelReason?,
    val data: AgendaEventData,
    val reminderOptions: ReminderOptions,
    val activeReminder: Reminder,
    val createdAt: Instant,
    private val domainEventRegistry: DomainEventRegistry<AgendaLifecycleEvent>
) : DomainEventRegistry<AgendaLifecycleEvent> by domainEventRegistry {

    companion object {
        fun create(
            id: String,
            timeSpan: TimeSpan,
            attendee: Attendee,
            data: AgendaEventData,
            reminderOptions: ReminderOptions,
        ): AgendaEvent {
            val agendaEvent =
                AgendaEvent(
                    id,
                    timeSpan,
                    attendee,
                    null,
                    data,
                    reminderOptions,
                    Reminder.from(timeSpan, reminderOptions),
                    Instant.now(),
                    DomainEventRegistry.delegate()
                )
            return agendaEvent.also { it.addEvent(AgendaEventScheduled(it)) }
        }
    }

    val title =
        when (data) {
            is BasicEventData -> data.title
            is AppointmentEventData -> data.services.joinToString(",") { it.name }
        }

    fun reschedule(timeSpan: TimeSpan): AgendaEvent {
        require(cancelReason == null) { "Event already canceled" }
        val reminder =
            activeReminder
                .rescheduleTimeToSent(timeSpan, reminderOptions)
                .fold(onSuccess = { it }, onFailure = { Reminder.from(timeSpan, reminderOptions) })
        return copy(timeSpan = timeSpan, activeReminder = reminder).also {
            it.addEvent(AgendaEventRescheduled(it))
        }
    }

    fun cancel(reason: CancelReason): AgendaEvent {
        require(cancelReason == null) { "Event already canceled" }
        return copy(cancelReason = reason).also { it.addEvent(AgendaEventDeleted(it)) }
    }

    fun updateReminderStatus(reminderStatus: ReminderStatus): AgendaEvent {
        return activeReminder
            ?.updateStatus(reminderStatus)
            ?.map { copy(activeReminder = it) }
            ?.getOrThrow()
            ?: this
    }
}
