package template

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
)

type Renderer struct {
	basePath string
}

func NewRenderer(basePath string) *Renderer {
	return &Renderer{basePath: basePath}
}

func (renderer *Renderer) Render(ctx context.Context, data application.CustomerNotificationTemplateData) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	if strings.TrimSpace(renderer.basePath) == "" {
		return "", fmt.Errorf("templates path is required")
	}
	path := filepath.Join(renderer.basePath, data.NotificationType, data.NotificationChannel+".body.tmpl")
	tpl, err := texttemplate.New(filepath.Base(path)).Funcs(templateFunctions()).ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("parse notification template %q: %w", path, err)
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data.Values); err != nil {
		return "", fmt.Errorf("execute notification template %q: %w", path, err)
	}
	return out.String(), nil
}

func templateFunctions() texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"dateFormat":         formatDateWithLayoutDefault,
		"dateFormatIn":       formatDateWithLayoutIn,
		"isChristmasHoliday": isChristmasHoliday,
	}
}
