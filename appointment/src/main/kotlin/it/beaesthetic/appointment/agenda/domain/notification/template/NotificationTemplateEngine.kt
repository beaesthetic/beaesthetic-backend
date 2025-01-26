package it.beaesthetic.appointment.agenda.domain.notification.template

import it.beaesthetic.appointment.agenda.domain.notification.Notification

class NotificationTemplateEngine(
    private val templates: List<NotificationTemplate>,
    private val fallbackTemplate: NotificationTemplate?
) {

    fun process(event: Notification): Result<String> {
        val template = templates.firstOrNull { it.isValid(event) } ?: fallbackTemplate
        if (template == null) {
            return Result.failure(IllegalArgumentException("No valid template found"))
        }
        return runCatching { template.invoke(event) }
            .recoverCatching {
                fallbackTemplate?.invoke(event) ?: throw it
            }
    }

    companion object {
        fun builder() = Builder()
        data class Builder(
            private var templates: List<NotificationTemplate> = emptyList(),
            private var fallbackTemplate: NotificationTemplate? = null
        ) {
            fun add(template: NotificationTemplate) = copy(templates = templates + template)
            fun fallback(template: NotificationTemplate) = copy(fallbackTemplate = template)

            fun build(): NotificationTemplateEngine {
                assert(templates.isNotEmpty()) { "Template is empty" }
                return NotificationTemplateEngine(templates, fallbackTemplate)
            }
        }
    }
}
