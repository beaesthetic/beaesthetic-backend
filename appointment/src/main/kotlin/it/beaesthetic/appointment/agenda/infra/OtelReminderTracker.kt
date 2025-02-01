package it.beaesthetic.appointment.agenda.infra

import io.micrometer.core.instrument.Gauge
import io.micrometer.core.instrument.MeterRegistry
import it.beaesthetic.appointment.agenda.application.reminder.ReminderTracker
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import jakarta.enterprise.context.ApplicationScoped
import java.util.concurrent.atomic.AtomicInteger

@ApplicationScoped
class OtelReminderTracker(meterRegistry: MeterRegistry) : ReminderTracker {

    private val reminderScheduled =
        meterRegistry.counter("reminders_scheduled_total", "type", "scheduled")
    private val reminderSentRequested =
        meterRegistry.counter("reminders_sent_requested_total", "type", "sent_requested")
    private val reminderSent = meterRegistry.counter("reminders_sent_total", "type", "sent")
    private val reminderFailed = meterRegistry.counter("reminders_failed_total", "type", "failed")
    private val reminderUnprocessable =
        meterRegistry.counter("reminders_unprocessable_total", "type", "unprocessable")
    private val reminderDeleted =
        meterRegistry.counter("reminders_deleted_total", "type", "deleted")

    private val reminderFailedCount = AtomicInteger(0)

    init {
        Gauge.builder("appointments.reminders.missing") { reminderFailedCount.toLong() }
            .baseUnit("appointments")
            .description("Numero potenziale di reminder non inviati")
            .register(meterRegistry)
    }

    override fun trackReminderState(reminderStatus: ReminderStatus) {
        when (reminderStatus) {
            ReminderStatus.PENDING,
            ReminderStatus.SCHEDULED -> reminderScheduled.increment()
            ReminderStatus.SENT_REQUESTED -> reminderSentRequested.increment()
            ReminderStatus.SENT -> reminderSent.increment()
            ReminderStatus.FAIL_TO_SEND -> reminderFailed.increment()
            ReminderStatus.UNPROCESSABLE -> reminderUnprocessable.increment()
            ReminderStatus.DELETED -> reminderDeleted.increment()
        }
    }

    override fun trackFailedReminders(failedReminder: Int) {
        reminderFailedCount.set(failedReminder)
    }
}
