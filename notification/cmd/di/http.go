package di

import (
	nethttp "net/http"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/port/health"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/port/http"
)

func (d *DiContainer) GetHttpServer() *nethttp.Server {
	return singleton(d, "httpServer", func() *nethttp.Server {
		ginEngine := http.New(d.GetHttpHandlers(), d.Config.App.Name)
		return &nethttp.Server{Addr: d.Config.HTTP.Addr, Handler: ginEngine}
	})
}

func (d *DiContainer) GetHttpHandlers() *http.HttpHandlers {
	return singleton(d, "httpHandlers", func() *http.HttpHandlers {
		return &http.HttpHandlers{
			SmsWebhook:    d.SmsWebhookHttpHandler(),
			HealthChecker: d.HealthCheckHandler(),
		}
	})
}

func (d *DiContainer) SmsWebhookHttpHandler() *http.Server {
	return singleton(d, "smsWebhookHttpHandler", func() *http.Server {
		return http.NewSmsWebhookHandler(d.GetCustomerNotificationService(), d.Log)
	})
}

func (d *DiContainer) HealthCheckHandler() health.HealthCheckHandler {
	return singleton(d, "healthCheckHandler", func() health.HealthCheckHandler {
		return health.NewHealthCheckHandler(d.GetPostgresDatabase())
	})
}
