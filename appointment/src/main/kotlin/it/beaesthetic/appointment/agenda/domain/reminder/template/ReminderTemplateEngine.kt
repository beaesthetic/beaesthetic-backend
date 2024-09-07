package it.beaesthetic.appointment.agenda.domain.reminder.template

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent

class ReminderTemplateEngine(
    private val templates: List<ReminderTemplate>,
    private val fallbackTemplate: ReminderTemplate
) {

    fun process(event: AgendaEvent): String {
        val template = templates.firstOrNull { it.isValid(event) } ?: fallbackTemplate
        return runCatching { template.invoke(event) }
            .recoverCatching { fallbackTemplate.invoke(event) }
            .getOrThrow()
    }

    companion object {
        fun builder() = Builder()
        data class Builder(
            private var templates: List<ReminderTemplate> = emptyList(),
            private var fallbackTemplate: ReminderTemplate? = null
        ) {
            fun add(template: ReminderTemplate) = copy(templates = templates + template)
            fun fallback(template: ReminderTemplate) = copy(fallbackTemplate = template)

            fun build(): ReminderTemplateEngine {
                assert(templates.isNotEmpty()) { "Template is empty" }
                assert(fallbackTemplate != null) { "FallbackTemplate is empty" }
                return ReminderTemplateEngine(templates, fallbackTemplate!!)
            }
        }
    }
}
