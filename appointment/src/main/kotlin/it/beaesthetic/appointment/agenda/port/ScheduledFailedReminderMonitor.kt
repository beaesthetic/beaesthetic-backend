package it.beaesthetic.appointment.agenda.port

import io.quarkus.scheduler.Scheduled
import it.beaesthetic.appointment.agenda.application.reminder.FailedReminderMonitor
import it.beaesthetic.appointment.agenda.domain.Clock
import jakarta.enterprise.context.ApplicationScoped
import java.time.Instant
import kotlin.time.Duration
import org.eclipse.microprofile.config.inject.ConfigProperty

@ApplicationScoped
class ScheduledFailedReminderMonitor(private val failedReminderMonitor: FailedReminderMonitor) {

    @ConfigProperty(name = "monitor.reminders.failed.frequency")
    private lateinit var frequency: String

    private val clock = Clock.default()

    private val shiftedClock by lazy {
        val shiftMillis = Duration.parse(frequency).inWholeMilliseconds
        object : Clock {
            override fun now(): Instant {
                return clock.now().minusMillis(shiftMillis)
            }
        }
    }

    @Scheduled(every = "\${monitor.reminders.failed.frequency}")
    suspend fun scheduledCheck() {
        failedReminderMonitor(shiftedClock)
    }
}
