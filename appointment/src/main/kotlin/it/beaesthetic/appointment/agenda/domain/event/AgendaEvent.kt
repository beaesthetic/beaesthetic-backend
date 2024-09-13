package it.beaesthetic.appointment.agenda.domain.event

import it.beaesthetic.appointment.agenda.domain.Clock
import it.beaesthetic.appointment.agenda.domain.reminder.Reminder
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import it.beaesthetic.appointment.common.DomainEventRegistry
import java.time.Duration
import java.time.Instant

@JvmInline value class AgendaEventId(val value: String)

data class Attendee(val id: String, val displayName: String)

data class AgendaEvent(
    val id: AgendaEventId,
    val timeSpan: TimeSpan,
    val attendee: Attendee,
    val cancelReason: CancelReason?,
    val data: AgendaEventData,
    val reminder: Reminder,
    val remindBefore: Duration,
    val createdAt: Instant,
    private val domainEventRegistry: DomainEventRegistry<AgendaLifecycleEvent>
) : DomainEventRegistry<AgendaLifecycleEvent> by domainEventRegistry {

    companion object {
        fun create(
            id: AgendaEventId,
            timeSpan: TimeSpan,
            attendee: Attendee,
            data: AgendaEventData,
            reminderBefore: Duration,
        ): AgendaEvent {
            val agendaEvent =
                AgendaEvent(
                    id,
                    timeSpan,
                    attendee,
                    null,
                    data,
                    Reminder(id, ReminderStatus.PENDING, null),
                    reminderBefore,
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
        return copy(timeSpan = timeSpan).also { it.addEvent(AgendaEventRescheduled(it)) }
    }

    fun cancel(reason: CancelReason): AgendaEvent {
        require(cancelReason == null) { "Event already canceled" }
        return copy(cancelReason = reason).also { it.addEvent(AgendaEventDeleted(it)) }
    }

    fun updateReminder(reminder: Reminder): AgendaEvent {
        return copy(reminder = reminder)
    }

    fun trackReminderAsSent(clock: Clock): AgendaEvent {
        return copy(reminder = reminder.copy(status = ReminderStatus.SENT, sentAt = clock.now()))
    }
}
