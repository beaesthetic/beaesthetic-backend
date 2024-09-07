package it.beaesthetic.appointment.agenda.domain.reminder

import it.beaesthetic.appointment.agenda.domain.event.TimeSpan
import java.time.Instant

enum class ReminderStatus {
    PENDING,
    SCHEDULED,
    SENT_REQUESTED,
    SENT,
    DELETED
}

data class Reminder(val status: ReminderStatus, val timeToSend: Instant) {

    private val allowedTransitions =
        mapOf(
            ReminderStatus.PENDING to ReminderStatus.entries,
            ReminderStatus.SCHEDULED to listOf(ReminderStatus.SENT_REQUESTED),
            ReminderStatus.SENT_REQUESTED to listOf(ReminderStatus.SENT),
        )

    companion object {
        fun from(timeSpan: TimeSpan, options: ReminderOptions) = from(timeSpan.start, options)
        fun from(start: Instant, options: ReminderOptions) =
            Reminder(
                status = ReminderStatus.PENDING,
                timeToSend = start - options.triggerBefore,
            )
    }

    val canBeRescheduled
        get() = status == ReminderStatus.SCHEDULED || status == ReminderStatus.PENDING

    fun rescheduleTimeToSent(
        timeSpan: TimeSpan,
        reminderOptions: ReminderOptions
    ): Result<Reminder> {
        return if (canBeRescheduled) {
            Result.success(copy(timeToSend = timeSpan.start - reminderOptions.triggerBefore))
        } else {
            Result.failure(IllegalStateException("Cannot rescheduled a Reminder already processed"))
        }
    }

    fun updateStatus(reminderStatus: ReminderStatus): Result<Reminder> {
        val allowedTransition = allowedTransitions[status]
        return if (
            reminderStatus == status || allowedTransition?.contains(reminderStatus) == true
        ) {
            Result.success(copy(status = reminderStatus))
        } else {
            Result.failure(
                IllegalStateException(
                    "Cannot move status from $status to $reminderStatus. " +
                        "AllowedTransitions: ${allowedTransition?.joinToString(",")}"
                )
            )
        }
    }
}
