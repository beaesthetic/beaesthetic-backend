package it.beaesthetic.appointment.agenda.port

import io.quarkus.runtime.StartupEvent
import it.beaesthetic.appointment.agenda.application.reminder.ScheduleReminder
import it.beaesthetic.appointment.agenda.application.reminder.ScheduleReminderHandler
import it.beaesthetic.appointment.agenda.domain.event.AgendaRepository
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import jakarta.enterprise.context.ApplicationScoped
import jakarta.enterprise.event.Observes
import java.time.Duration
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import kotlinx.coroutines.runBlocking
import org.eclipse.microprofile.config.inject.ConfigProperty
import org.jboss.logging.Logger

@ApplicationScoped
class RecoveryDelayTrigger(
    private val agendaRepository: AgendaRepository,
    private val scheduleReminderHandler: ScheduleReminderHandler,
    @ConfigProperty(name = "recovery.enabled") private val isRecoveryEnabled: Boolean
) {

    private val log = Logger.getLogger(RecoveryDelayTrigger::class.java)

    companion object {
        private val RECOVER_DELAY = Duration.ofSeconds(2)
    }

    fun started(@Observes event: StartupEvent) {
        if (isRecoveryEnabled) {
            log.info("Started Recovery DelayTrigger")
            Executors.newSingleThreadScheduledExecutor()
                .schedule(
                    { runBlocking { recoverStuckAgendaReminder() } },
                    RECOVER_DELAY.toSeconds(),
                    TimeUnit.SECONDS
                )
        } else {
            log.warn("Initial agenda event recovery is disabled")
        }
    }

    private suspend fun recoverStuckAgendaReminder() {
        val stuckEvents = agendaRepository.findEventsWithReminderState(ReminderStatus.PENDING)
        log.info("Find ${stuckEvents.size} stuck agenda events. Trying to schedule reminder")
        stuckEvents.forEach { event ->
            scheduleReminderHandler.handle(ScheduleReminder(event.id)).onFailure {
                log.error("Failed to recover agenda event ${event.id}", it)
            }
        }
    }
}
