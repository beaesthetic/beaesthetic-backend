package it.beaesthetic.appointment.agenda

import it.beaesthetic.appointment.agenda.domain.notification.NotificationType
import it.beaesthetic.appointment.agenda.domain.notification.template.NotificationTemplate.Companion.template
import it.beaesthetic.appointment.agenda.domain.notification.template.NotificationTemplateEngine
import it.beaesthetic.appointment.agenda.domain.notification.template.TemplateRuleUtils.formatDateToRome
import it.beaesthetic.appointment.agenda.domain.notification.template.TemplateRuleUtils.formatTime
import it.beaesthetic.appointment.agenda.domain.notification.template.TemplateRuleUtils.timeIncludeDayRange
import jakarta.enterprise.context.Dependent
import jakarta.enterprise.inject.Produces
import jakarta.inject.Singleton
import java.time.MonthDay

@Dependent
object NotificationTemplateConfig {

    @Produces
    @Singleton
    fun notificationTemplateEngine(): NotificationTemplateEngine {
        // TODO: think about ntofiication variation concept to handle somehting like cristhmas
        return NotificationTemplateEngine.builder()
            .add(confirmRescheduled)
            .add(confirmTemplate)
            .add(christmasReminderTemplate)
            .add(defaultReminderTemplate)
            .build()
    }

    private val christmasReminderTemplate = template {
        whenCondition { it.type is NotificationType.Reminder }
        whenCondition {
            timeIncludeDayRange(MonthDay.of(12, 10), MonthDay.of(1, 6))
                .invoke(it.event.timeSpan.start)
        }
        apply {
            val day = formatDateToRome(it.event.timeSpan.start)
            val hour = formatTime(it.event.timeSpan.start)
            """
                    Il centro Be Aesthetic ti ricorda il tuo appuntamento di $day alle ore $hour.
                    Buona giornata e buone feste!
                """
                .trimIndent()
        }
    }

    private val confirmTemplate = template {
        whenCondition { it.type is NotificationType.Confirmation }
        apply {
            val day = formatDateToRome(it.event.timeSpan.start)
            val hour = formatTime(it.event.timeSpan.start)
            """
                Il centro Be Aesthetic ti conferma la prenotazione del tuo appuntamento per il giorno $day alle ore $hour.
                Buona giornata!
            """
                .trimIndent()
        }
    }

    private val confirmRescheduled = template {
        whenCondition { it.type is NotificationType.Confirmation && it.type.isRescheduled }
        apply {
            val day = formatDateToRome(it.event.timeSpan.start)
            val hour = formatTime(it.event.timeSpan.start)
            """
                 Il centro Be Aesthetic ti informa che il tuo appuntamento è stato spostato. La nuova data è $day alle ore $hour.
                 Buona giornata!
            """
                .trimIndent()
        }
    }

    private val defaultReminderTemplate = template {
        whenCondition { it.type is NotificationType.Reminder }
        apply {
            val day = formatDateToRome(it.event.timeSpan.start)
            val hour = formatTime(it.event.timeSpan.start)
            """
                    Il centro Be Aesthetic ti ricorda il tuo appuntamento di $day alle ore $hour.
                    Buona giornata!
                """
                .trimIndent()
        }
    }
}
