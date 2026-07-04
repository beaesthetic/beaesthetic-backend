package di

import (
	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/port/health"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/port/http"
)

func (d *DiContainer) GetGinHttpServer() *gin.Engine {
	return singleton(d, "httpServer", func() *gin.Engine {
		return http.New(d.GetHttpHandlers())
	})
}

func (d *DiContainer) GetHttpHandlers() *http.HttpHandlers {
	return singleton(d, "httpHandlers", func() *http.HttpHandlers {
		return &http.HttpHandlers{
			Sms:           d.SmsHttpHandler(),
			SmsWebhook:    d.SmsWebhookHttpHandler(),
			HealthChecker: d.HealthCheckHandler(),
		}
	})
}

func (d *DiContainer) SmsHttpHandler() *http.Server {
	return singleton(d, "smsHttpHandler", func() *http.Server {
		return http.NewSmsHandler(d.GetNotificationService(), d.Log)
	})
}

func (d *DiContainer) SmsWebhookHttpHandler() *http.Server {
	return singleton(d, "smsWebhookHttpHandler", func() *http.Server {
		return http.NewSmsHandler(d.GetNotificationService(), d.Log)
	})
}

func (d *DiContainer) HealthCheckHandler() health.HealthCheckHandler {
	return singleton(d, "healthCheckHandler", func() health.HealthCheckHandler {
		return health.NewHealthCheckHandler(d.GetPostgresDatabase())
	})
}
