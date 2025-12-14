package it.beaesthetic.appointment.agenda.domain.notification

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent

// TODO: refactor this. A notification should be composed onlu by text and destination not the
// entire event
// THIS means move out the template engine from notification service. And create a notification
// from template engine

sealed interface NotificationType {
    data object Reminder : NotificationType

    data class Confirmation(val isRescheduled: Boolean = false) : NotificationType
}

data class Notification(val type: NotificationType, val event: AgendaEvent) {
    companion object {
        fun reminder(event: AgendaEvent): Notification {
            return Notification(NotificationType.Reminder, event)
        }

        fun confirmation(event: AgendaEvent, isRescheduled: Boolean = false): Notification {
            return Notification(NotificationType.Confirmation(isRescheduled), event)
        }
    }
}
