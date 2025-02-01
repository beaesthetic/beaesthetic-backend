package it.beaesthetic.appointment.agenda.application.reminder

import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.domain.Clock
import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.reminder.ReminderStatus
import it.beaesthetic.appointment.agenda.infra.mongo.EntityMapper
import it.beaesthetic.appointment.agenda.infra.mongo.PanacheAgendaRepository
import jakarta.enterprise.context.ApplicationScoped
import java.time.LocalDate
import java.time.LocalDateTime
import java.time.ZoneId
import org.bson.Document
import org.jboss.logging.Logger

@ApplicationScoped
class FailedReminderMonitor(
    private val panacheAgendaRepository: PanacheAgendaRepository,
    private val reminderTracker: ReminderTracker
) {
    private val log = Logger.getLogger(FailedReminderMonitor::class.java)

    companion object {
        private val expectedStatus =
            setOf(
                ReminderStatus.DELETED,
                ReminderStatus.SENT,
                ReminderStatus.UNPROCESSABLE,
                ReminderStatus.PENDING,
            )
    }

    suspend operator fun invoke(clock: Clock) {
        log.info("Trigger failed reminder monitor. Started")
        val potentialFailedReminders =
            findTomorrowEventsUntilNow(clock).count { !expectedStatus.contains(it.reminder.status) }

        log.info("Found $potentialFailedReminders potential failed reminders. End")
        reminderTracker.trackFailedReminders(potentialFailedReminders)
    }

    private suspend fun findTomorrowEventsUntilNow(clock: Clock): List<AgendaEvent> {
        val zoneId = ZoneId.systemDefault()
        val now = clock.now()
        val tomorrowSameTime =
            LocalDateTime.of(
                    LocalDate.ofInstant(now, zoneId).plusDays(1),
                    LocalDateTime.ofInstant(now, zoneId).toLocalTime()
                )
                .atZone(zoneId)
                .toInstant()

        log.info("Monitor failed reminder from: $now, until: $tomorrowSameTime")

        val query =
            Document("start", Document("\$gte", now).append("\$lte", tomorrowSameTime))
                .append("isCancelled", false)
        return panacheAgendaRepository
            .find(query, Document("start", 1))
            .stream()
            .map { EntityMapper.toDomain(it) }
            .collect()
            .asList()
            .awaitSuspending()
    }
}
