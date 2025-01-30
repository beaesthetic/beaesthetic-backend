package it.beaesthetic.appointment.agenda.domain.reminder

import it.beaesthetic.appointment.agenda.application.reminder.ReminderTracker
import it.beaesthetic.appointment.agenda.domain.Clock
import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.notification.Notification
import it.beaesthetic.appointment.agenda.domain.notification.NotificationService
import it.beaesthetic.appointment.agenda.domain.notification.NotificationType
import java.time.Duration
import java.time.Instant
import org.jboss.logging.Logger

class ReminderService(
    private val clock: Clock,
    private val reminderOptions: ReminderOptions,
    private val reminderScheduler: ReminderScheduler,
    private val notificationService: NotificationService,
    private val reminderTracker: ReminderTracker,
) {

    private val log = Logger.getLogger(ReminderService::class.java)

    suspend fun scheduleReminder(event: AgendaEvent): Result<AgendaEvent> {
        val sendAt =
            computeSendTime(
                clock,
                event.timeSpan.start,
                event.remindBefore,
                reminderOptions.noSendThreshold,
                reminderOptions.immediateSendThreshold
            )
        log.info("Reminder for ${event.id}, start ${event.timeSpan.start}, scheduled at $sendAt")
        val updatedReminder =
            when {
                sendAt == null ->
                    Result.success(event.reminder.copy(status = ReminderStatus.UNPROCESSABLE))
                else ->
                    runCatching { reminderScheduler.scheduleReminder(event.reminder, sendAt) }
                        .map { event.reminder.copy(status = ReminderStatus.SCHEDULED) }
            }

        return updatedReminder
            .onSuccess { reminderTracker.trackReminderState(it.status) }
            .map { event.updateReminder(it) }
    }

    suspend fun unscheduleReminder(event: AgendaEvent): Result<AgendaEvent> {
        // fix change remind unschedule interface
        return kotlin
            .runCatching {
                reminderScheduler.unschedule(event.reminder).copy(status = ReminderStatus.DELETED)
            }
            .onSuccess { reminderTracker.trackReminderState(it.status) }
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
                    if (sendAt == null && event.cancelReason == null) {
                        event.reminder.copy(status = ReminderStatus.UNPROCESSABLE)
                    } else {
                        runCatching {
                                notificationService.trackAndSendNotification(
                                    Notification(NotificationType.Reminder, event),
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
        }.apply { reminderTracker.trackReminderState(reminder.status) }
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
