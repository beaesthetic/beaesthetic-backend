package it.beaesthetic.appointment.agenda.domain.reminder.template

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent

interface ReminderTemplate {
    fun invoke(event: AgendaEvent): String
    fun isValid(event: AgendaEvent): Boolean

    companion object {
        fun from(
            applyWhen: (event: AgendaEvent) -> Boolean,
            apply: (event: AgendaEvent) -> String
        ) =
            object : ReminderTemplate {
                override fun invoke(event: AgendaEvent) = apply(event)
                override fun isValid(event: AgendaEvent) = applyWhen(event)
            }

        fun template(block: ReminderTemplateBuilder.() -> Unit): ReminderTemplate {
            val builder = ReminderTemplateBuilder()
            builder.block()
            return builder.build()
        }
    }
}
