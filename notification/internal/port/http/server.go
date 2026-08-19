package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api/smswebhook"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/port/health"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type HttpHandlers struct {
	SmsWebhook    smswebhook.StrictServerInterface
	HealthChecker health.HealthCheckHandler
}

func New(handlers *HttpHandlers, serviceName string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware(serviceName, otelgin.WithFilter(func(request *nethttp.Request) bool {
		return request.URL.Path != "/health"
	})))
	smswebhook.RegisterHandlers(router, smswebhook.NewStrictHandlerWithOptions(handlers.SmsWebhook, nil, smswebhook.StrictGinServerOptions{
		HandlerErrorFunc: strictErrorHandler,
	}))

	router.GET("/health", handlers.HealthChecker)

	return router
}
