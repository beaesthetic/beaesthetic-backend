package it.beaesthetic.appointment.agenda

import it.beaesthetic.appointment.agenda.domain.reminder.template.ReminderTemplate.Companion.template
import it.beaesthetic.appointment.agenda.domain.reminder.template.ReminderTemplateEngine
import it.beaesthetic.appointment.agenda.domain.reminder.template.TemplateRuleUtils.formatDateToRome
import it.beaesthetic.appointment.agenda.domain.reminder.template.TemplateRuleUtils.timeIncludeDayRange
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces
import jakarta.inject.Singleton
import java.time.MonthDay

@Dependent
object ReminderTemplateConfig {

    @Produces
    @Singleton
    fun reminderTemplateEngine(): ReminderTemplateEngine {
        return ReminderTemplateEngine.builder()
            .add(christmasTemplate)
            .fallback(defaultTemplate)
            .build()
    }

    private val christmasTemplate = template {
        whenCondition {
            timeIncludeDayRange(MonthDay.of(12, 10), MonthDay.of(1, 6)).invoke(it.timeSpan.start)
        }
        apply {
            val start = formatDateToRome(it.timeSpan.start)
            val end = formatDateToRome(it.timeSpan.end)
            """
                    Il centro Be Aesthetic ti ricorda il tuo appuntamento di $start alle ore ${end}.
                    Buona giornata e buone feste!
                """
                .trimIndent()
        }
    }

    private val defaultTemplate = template {
        whenCondition { true }
        apply {
            val start = formatDateToRome(it.timeSpan.start)
            val end = formatDateToRome(it.timeSpan.end)
            """
                    Il centro Be Aesthetic ti ricorda il tuo appuntamento di $start alle ore ${end}.
                    Buona giornata!
                """
                .trimIndent()
        }
    }
}
