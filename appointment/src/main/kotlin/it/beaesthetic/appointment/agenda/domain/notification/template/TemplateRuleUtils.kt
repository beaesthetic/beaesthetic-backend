package it.beaesthetic.appointment.agenda.domain.notification.template

import java.time.Instant
import java.time.MonthDay
import java.time.ZoneId
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter
import java.util.*

object TemplateRuleUtils {
    fun timeIncludeDayRange(start: MonthDay, end: MonthDay): (Instant) -> Boolean {
        return { instant: Instant ->
            val localDate = instant.atOffset(ZoneOffset.UTC).toLocalDate()
            val currentMonthDay = MonthDay.from(localDate)

            if (start.isBefore(end)) {
                // If the range doesn't span across years, like March 10 to June 5
                currentMonthDay >= start && currentMonthDay <= end
            } else {
                // If the range spans across years, like December 10 to January 6
                currentMonthDay >= start || currentMonthDay <= end
            }
        }
    }

    fun formatDateToRome(instant: Instant): String {
        val zone = ZoneId.of("Europe/Rome")
        val locale = Locale.ITALY
        val formatter = DateTimeFormatter.ofPattern("EEEE d MMMM, yyyy", locale)

        return instant.atZone(zone).format(formatter)
    }

    fun formatTime(instant: Instant): String {
        val zone = ZoneId.of("Europe/Rome")
        val locale = Locale.ITALY
        val formatter = DateTimeFormatter.ofPattern("HH:mm", locale)

        return instant.atZone(zone).format(formatter)
    }
}
