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
