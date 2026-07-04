package application

import (
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

func TestReminderNotificationPayloadUsesDefaultItalianTemplate(t *testing.T) {
	event := notificationTemplateEvent(time.Date(2026, time.July, 4, 13, 30, 0, 0, time.UTC))
	title, content := ReminderNotificationPayload(event)

	if title != "" {
		t.Fatalf("title=%q want empty", title)
	}
	want := "Il centro Be Aesthetic ti ricorda il tuo appuntamento di sabato 4 luglio, 2026 alle ore 15:30.\nBuona giornata!"
	if content != want {
		t.Fatalf("content=%q want=%q", content, want)
	}
}

func TestReminderNotificationPayloadUsesChristmasTemplate(t *testing.T) {
	event := notificationTemplateEvent(time.Date(2026, time.December, 10, 8, 0, 0, 0, time.UTC))
	_, content := ReminderNotificationPayload(event)

	want := "Il centro Be Aesthetic ti ricorda il tuo appuntamento di giovedì 10 dicembre, 2026 alle ore 09:00.\nBuona giornata e buone feste!"
	if content != want {
		t.Fatalf("content=%q want=%q", content, want)
	}
}

func TestConfirmationNotificationPayloadUsesItalianTemplate(t *testing.T) {
	event := notificationTemplateEvent(time.Date(2026, time.July, 4, 13, 30, 0, 0, time.UTC))
	title, content := confirmationNotificationPayload(event, false)

	if title != "" {
		t.Fatalf("title=%q want empty", title)
	}
	want := "Il centro Be Aesthetic ti conferma la prenotazione del tuo appuntamento per il giorno sabato 4 luglio, 2026 alle ore 15:30.\nBuona giornata!"
	if content != want {
		t.Fatalf("content=%q want=%q", content, want)
	}
}

func TestConfirmationNotificationPayloadUsesRescheduledItalianTemplate(t *testing.T) {
	event := notificationTemplateEvent(time.Date(2026, time.July, 4, 13, 30, 0, 0, time.UTC))
	_, content := confirmationNotificationPayload(event, true)

	want := "Il centro Be Aesthetic ti informa che il tuo appuntamento è stato spostato. La nuova data è sabato 4 luglio, 2026 alle ore 15:30.\nBuona giornata!"
	if content != want {
		t.Fatalf("content=%q want=%q", content, want)
	}
}

func notificationTemplateEvent(start time.Time) *domain.AgendaEvent {
	return &domain.AgendaEvent{
		ID:           "event-id",
		Title:        "Consulto",
		Start:        start,
		End:          start.Add(time.Hour),
		Attendee:     domain.Attendee{ID: "customer-id", DisplayName: "Customer"},
		RemindBefore: 24 * time.Hour,
	}
}
