package it.beaesthetic.appointment.agenda.domain.notification.template

import it.beaesthetic.appointment.agenda.domain.notification.Notification

class NotificationTemplateBuilder {
    private var applyWhen: List<(Notification) -> Boolean> = emptyList()
    private var apply: (Notification) -> String = { "" }

    fun whenCondition(condition: (Notification) -> Boolean) {
        applyWhen += condition
    }

    fun apply(action: (Notification) -> String) {
        apply = action
    }

    fun build(): NotificationTemplate {
        return NotificationTemplate.from({e -> applyWhen.all { it(e) }}, apply)
    }
}
