package it.beaesthetic.appointment.agenda.domain.reminder.template

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent

class ReminderTemplateBuilder {
    private var applyWhen: (AgendaEvent) -> Boolean = { true }
    private var apply: (AgendaEvent) -> String = { "" }

    fun whenCondition(condition: (AgendaEvent) -> Boolean) {
        applyWhen = condition
    }

    fun apply(action: (AgendaEvent) -> String) {
        apply = action
    }

    fun build(): ReminderTemplate {
        return ReminderTemplate.from(applyWhen, apply)
    }
}
