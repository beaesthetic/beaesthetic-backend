package it.beaesthetic.appointment.agenda.application.events

import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.appointment.agenda.application.notification.SendNotificationHandler
import it.beaesthetic.appointment.agenda.config.NotificationConfiguration
import it.beaesthetic.appointment.agenda.domain.event.AgendaEvent
import it.beaesthetic.appointment.agenda.domain.event.AgendaEventScheduled
import it.beaesthetic.appointment.agenda.domain.event.Attendee
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.mock

class AgendaEventToNotificationPolicyTest {

    private val notificationConfiguration = object : NotificationConfiguration {
        override fun whitelist(): List<String> = listOf("123")
    }

    private lateinit var sendNotificationHandler: SendNotificationHandler
    private lateinit var agentEventToNotificationPolicy: AgendaEventToNotificationPolicy

    @BeforeEach
    fun setup() {
        sendNotificationHandler = mock()
        agentEventToNotificationPolicy = AgendaEventToNotificationPolicy(
            notificationConfiguration,
            sendNotificationHandler
        )
    }

    @Test
    fun aa(): Unit = runBlocking {
        val attendee1 = Attendee("123", "")
        val agendaEvent = mock<AgendaEvent> {
            on { attendee }.thenReturn(attendee1)
        }
        val event = AgendaEventScheduled(agendaEvent)
        agentEventToNotificationPolicy.handle(event)
            .awaitSuspending()
    }

}