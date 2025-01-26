package it.beaesthetic.appointment.agenda.domain.notification.template

import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.notification.Notification
import it.beaesthetic.appointment.agenda.domain.notification.NotificationType
import it.beaesthetic.appointment.agenda.domain.notification.template.NotificationTemplate.Companion.template
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock

class NotificationTemplateEngineTest {


    @Test
    fun `should process event with valid template`() {
        // Prepare a valid event and template
        val notificationTemplate1 = template {
            whenCondition { it.type is NotificationType.Reminder }
            apply { "TEST1" }
        }

        val notificationTemplate2 = template {
            whenCondition { it.type is NotificationType.Confirmation }
            apply { "TEST2" }
        }

        val engine = NotificationTemplateEngine.builder()
            .add(notificationTemplate2)
            .add(notificationTemplate1)
            .build()

        val agendaEvent = mock<AgendaEvent>()

        // Process the event
        val result = engine.process(Notification.reminder(agendaEvent))

        // Assertions
        assert(result.isSuccess)
        assertEquals("TEST1", result.getOrNull())
    }

    @Test
    fun `should return failure if no valid template and no fallback`() {
        // Prepare event and templates
        val confirmEvent = Notification.confirmation(mock(), true)
        val notificationTemplate = template {
            whenCondition { it.type is NotificationType.Reminder }
            apply { "TEST" }
        }

        val engine = NotificationTemplateEngine.builder()
            .add(notificationTemplate)
            .build()

        // Process the event with no matching template and no fallback
        val result = engine.process(confirmEvent)

        // Assertions
        assertTrue(result.isFailure)
        assertTrue(result.exceptionOrNull() is IllegalArgumentException)
        assertEquals("No valid template found", result.exceptionOrNull()?.message)
    }
}