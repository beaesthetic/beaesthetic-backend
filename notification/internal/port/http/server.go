package http

import (
	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api/smswebhook"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/port/health"
)

type HttpHandlers struct {
	SmsWebhook    smswebhook.StrictServerInterface
	HealthChecker health.HealthCheckHandler
}

func New(handlers *HttpHandlers) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	smswebhook.RegisterHandlers(router, smswebhook.NewStrictHandlerWithOptions(handlers.SmsWebhook, nil, smswebhook.StrictGinServerOptions{
		HandlerErrorFunc: strictErrorHandler,
	}))

	router.GET("/health", handlers.HealthChecker)

	return router
}
