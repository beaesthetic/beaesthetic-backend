package it.beaesthetic.appointment.agenda.domain.notification.template

import it.beaesthetic.appointment.agenda.domain.notification.Notification

interface NotificationTemplate {
    fun invoke(event: Notification): String

    fun isValid(event: Notification): Boolean

    companion object {
        fun from(
            applyWhen: (event: Notification) -> Boolean,
            apply: (event: Notification) -> String,
        ) =
            object : NotificationTemplate {
                override fun invoke(event: Notification) = apply(event)

                override fun isValid(event: Notification) = applyWhen(event)
            }

        fun template(block: NotificationTemplateBuilder.() -> Unit): NotificationTemplate {
            val builder = NotificationTemplateBuilder()
            builder.block()
            return builder.build()
        }
    }
}
