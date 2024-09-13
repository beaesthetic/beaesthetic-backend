package it.beaesthetic.appointment.agenda.domain.reminder

import it.beaesthetic.appointment.agenda.domain.Clock
import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.notification.NotificationService
import it.beaesthetic.appointment.agenda.domain.reminder.template.ReminderTemplateEngine
import java.time.Duration
import java.time.Instant

class ReminderService(
    private val clock: Clock,
    private val reminderOptions: ReminderOptions,
    private val reminderScheduler: ReminderScheduler,
    private val notificationService: NotificationService,
    private val templateEngine: ReminderTemplateEngine
) {

    suspend fun scheduleReminder(event: AgendaEvent): Result<AgendaEvent> {
        val sendAt =
            computeSendTime(
                clock,
                event.timeSpan.start,
                event.remindBefore,
                reminderOptions.noSendThreshold,
                reminderOptions.immediateSendThreshold
            )
        val updatedReminder =
            when {
                sendAt == null ->
                    Result.success(event.reminder.copy(status = ReminderStatus.UNPROCESSABLE))
                else ->
                    runCatching { reminderScheduler.scheduleReminder(event.reminder, sendAt) }
                        .map { event.reminder.copy(status = ReminderStatus.SCHEDULED) }
            }

        return updatedReminder.map { event.updateReminder(it) }
    }

    suspend fun unscheduleReminder(event: AgendaEvent): Result<AgendaEvent> {
        // fix change remind unschedule interface
        return kotlin
            .runCatching {
                reminderScheduler.unschedule(event.reminder).copy(status = ReminderStatus.DELETED)
            }
            .map { event.updateReminder(it) }
    }

    suspend fun sendReminder(event: AgendaEvent, phoneNumber: String): AgendaEvent {
        return when (event.reminder.status) {
            ReminderStatus.SCHEDULED -> {
                val sendAt =
                    computeSendTime(
                        clock,
                        event.timeSpan.start,
                        event.remindBefore,
                        reminderOptions.noSendThreshold,
                        reminderOptions.immediateSendThreshold
                    )
                val updatedReminder =
                    if (sendAt == null) {
                        event.reminder.copy(status = ReminderStatus.UNPROCESSABLE)
                    } else {
                        runCatching {
                                notificationService.trackAndSendReminderNotification(
                                    event,
                                    templateEngine,
                                    phoneNumber
                                )
                            }
                            .fold(
                                onSuccess = {
                                    event.reminder.copy(status = ReminderStatus.SENT_REQUESTED)
                                },
                                onFailure = { e ->
                                    event.reminder.copy(status = ReminderStatus.FAIL_TO_SEND)
                                }
                            )
                    }
                event.updateReminder(updatedReminder)
            }
            else -> event
        }
    }

    private fun computeSendTime(
        clock: Clock,
        eventAt: Instant,
        sendBefore: Duration,
        noSendThreshold: Duration,
        immediateSendThreshold: Duration
    ): Instant? {
        val now = clock.now()
        val potentialRemindDate = eventAt - sendBefore
        val millisFromNow =
            Duration.ofMillis(now.toEpochMilli() - potentialRemindDate.toEpochMilli())
        val delta = sendBefore - millisFromNow
        return when {
            delta < noSendThreshold -> null
            delta >= sendBefore -> potentialRemindDate
            else -> now + immediateSendThreshold
        }
    }
}
