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
	appointments *application.AppointmentService
	services     *application.ServiceService
	log          *zap.Logger
}

func NewServer(appointments *application.AppointmentService, services *application.ServiceService, log *zap.Logger) *Server {
	return &Server{appointments: appointments, services: services, log: log}
}

var _ httpserver.StrictServerInterface = (*Server)(nil)
