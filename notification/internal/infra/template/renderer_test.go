package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
)

func TestRendererExecutesFlatValuesMap(t *testing.T) {
	basePath := t.TempDir()
	templateDir := filepath.Join(basePath, "appointment_reminder")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "sms.body.tmpl"), []byte(`Ciao {{ .name }}, appuntamento {{ .date }}`), 0600); err != nil {
		t.Fatal(err)
	}

	content, err := NewRenderer(basePath).Render(context.Background(), application.CustomerNotificationTemplateData{
		NotificationType:    "appointment_reminder",
		NotificationChannel: "sms",
		Values: map[string]any{
			"name": "Ada",
			"date": "2026-07-20",
		},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if content != "Ciao Ada, appuntamento 2026-07-20" {
		t.Fatalf("content = %q", content)
	}
}

func TestRendererRendersBuiltInAppointmentTemplates(t *testing.T) {
	tests := []struct {
		name             string
		notificationType string
		startAt          string
		want             string
	}{
		{
			name:             "confirmation",
			notificationType: "appointment_confirmation",
			startAt:          "2026-07-04T13:30:00Z",
			want:             "Il centro Be Aesthetic ti conferma la prenotazione del tuo appuntamento per il giorno sabato 4 luglio, 2026 alle ore 15:30.\nBuona giornata!\n",
		},
		{
			name:             "reminder",
			notificationType: "appointment_reminder",
			startAt:          "2026-07-04T13:30:00Z",
			want:             "Il centro Be Aesthetic ti ricorda il tuo appuntamento di sabato 4 luglio, 2026 alle ore 15:30.\nBuona giornata!\n",
		},
		{
			name:             "rescheduled",
			notificationType: "appointment_rescheduled",
			startAt:          "2026-07-04T13:30:00Z",
			want:             "Il centro Be Aesthetic ti informa che il tuo appuntamento è stato spostato. La nuova data è sabato 4 luglio, 2026 alle ore 15:30.\nBuona giornata!\n",
		},
		{
			name:             "reminder christmas holidays",
			notificationType: "appointment_reminder",
			startAt:          "2026-12-10T08:00:00Z",
			want:             "Il centro Be Aesthetic ti ricorda il tuo appuntamento di giovedì 10 dicembre, 2026 alle ore 09:00.\nBuona giornata e buone feste!\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := NewRenderer("../../../templates").Render(context.Background(), application.CustomerNotificationTemplateData{
				NotificationType:    tt.notificationType,
				NotificationChannel: "sms",
				Values: map[string]any{
					"startAt": tt.startAt,
				},
			})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if content != tt.want {
				t.Fatalf("content = %q, want %q", content, tt.want)
			}
		})
	}
}

func TestRendererFormatsDateAndTimeWithDateFormat(t *testing.T) {
	basePath := t.TempDir()
	templateDir := filepath.Join(basePath, "appointment_reminder")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "sms.body.tmpl"), []byte(`{{ dateFormat "Monday 2 January" .startAt }} {{ dateFormat "15:04" .startAt }}`), 0600); err != nil {
		t.Fatal(err)
	}

	content, err := NewRenderer(basePath).Render(context.Background(), application.CustomerNotificationTemplateData{
		NotificationType:    "appointment_reminder",
		NotificationChannel: "sms",
		Values: map[string]any{
			"startAt": "2026-07-04T13:30:00Z",
		},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if content != "sabato 4 luglio 15:30" {
		t.Fatalf("content = %q", content)
	}
}

func TestRendererChecksChristmasHoliday(t *testing.T) {
	basePath := t.TempDir()
	templateDir := filepath.Join(basePath, "appointment_reminder")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "sms.body.tmpl"), []byte(`{{ if isChristmasHoliday .startAt }}holiday{{ else }}default{{ end }}`), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		startAt string
		want    string
	}{
		{name: "before range", startAt: "2026-12-07T12:00:00Z", want: "default"},
		{name: "rome start boundary", startAt: "2026-12-07T23:30:00Z", want: "holiday"},
		{name: "start boundary", startAt: "2026-12-08T12:00:00Z", want: "holiday"},
		{name: "end boundary", startAt: "2027-01-06T12:00:00Z", want: "holiday"},
		{name: "after range", startAt: "2027-01-07T12:00:00Z", want: "default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := NewRenderer(basePath).Render(context.Background(), application.CustomerNotificationTemplateData{
				NotificationType:    "appointment_reminder",
				NotificationChannel: "sms",
				Values: map[string]any{
					"startAt": tt.startAt,
				},
			})
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if content != tt.want {
				t.Fatalf("content = %q, want %q", content, tt.want)
			}
		})
	}
}

func TestRendererFormatsDateWithTemplateLayout(t *testing.T) {
	basePath := t.TempDir()
	templateDir := filepath.Join(basePath, "appointment_reminder")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "sms.body.tmpl"), []byte(`{{ dateFormat "Monday 2 January 2006" .startAt }}|{{ dateFormatIn "en_US" "Europe/Rome" "January 2 2006" .startAt }}`), 0600); err != nil {
		t.Fatal(err)
	}

	content, err := NewRenderer(basePath).Render(context.Background(), application.CustomerNotificationTemplateData{
		NotificationType:    "appointment_reminder",
		NotificationChannel: "sms",
		Values: map[string]any{
			"startAt": "2026-07-04T13:30:00Z",
		},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if content != "sabato 4 luglio 2026|July 4 2026" {
		t.Fatalf("content = %q", content)
	}
}
