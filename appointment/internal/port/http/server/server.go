package server

import (
	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/health"
	httpserver "github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/server/generated"
	"go.uber.org/zap"
)

type HttpHandlers struct {
	Appointment   httpserver.StrictServerInterface
	HealthChecker health.HealthCheckHandler
}

func New(handlers *HttpHandlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	httpserver.RegisterHandlers(r, httpserver.NewStrictHandler(handlers.Appointment, nil))
	r.GET("/health", handlers.HealthChecker)
	return r
}

type Server struct {
	service *application.Service
	log     *zap.Logger
}

func NewServer(service *application.Service, log *zap.Logger) *Server {
	return &Server{service: service, log: log}
}

var _ httpserver.StrictServerInterface = (*Server)(nil)
