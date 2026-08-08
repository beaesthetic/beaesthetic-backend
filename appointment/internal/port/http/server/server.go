package server

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	applicationv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/application/v2"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/health"
	httpserver "github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/server/generated"
	"go.uber.org/zap"
)

type HttpHandlers struct {
	Appointment   httpserver.StrictServerInterface
	Calendar      *Server
	HealthChecker health.HealthCheckHandler
}

func New(handlers *HttpHandlers, log *zap.Logger) *gin.Engine {
	if log == nil {
		log = zap.NewNop()
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginErrorLogger(log))
	httpserver.RegisterHandlers(r, httpserver.NewStrictHandlerWithOptions(handlers.Appointment, nil, httpserver.StrictGinServerOptions{
		RequestErrorHandlerFunc: func(ctx *gin.Context, err error) {
			log.Warn("http request error", zap.Error(err), zap.String("method", ctx.Request.Method), zap.String("path", ctx.FullPath()))
			ctx.JSON(400, gin.H{"msg": err.Error()})
		},
		ResponseErrorHandlerFunc: func(ctx *gin.Context, err error) {
			log.Error("http response error", zap.Error(err), zap.String("method", ctx.Request.Method), zap.String("path", ctx.FullPath()))
			ctx.JSON(500, gin.H{"msg": err.Error()})
		},
	}))
	if handlers.Calendar != nil {
		registerCalendarProtoRoutes(r, handlers.Calendar)
	}
	r.GET("/health", handlers.HealthChecker)
	return r
}

func ginErrorLogger(log *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		fields := []zap.Field{
			zap.String("method", ctx.Request.Method),
			zap.String("path", ctx.FullPath()),
			zap.String("raw_path", ctx.Request.URL.Path),
			zap.Int("status", ctx.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		}
		if len(ctx.Errors) > 0 {
			fields = append(fields, zap.String("gin_errors", ctx.Errors.String()))
			log.Error("http request failed", fields...)
			return
		}
		if ctx.Writer.Status() >= 500 {
			log.Error("http request failed", fields...)
			return
		}
		if ctx.Writer.Status() >= 400 {
			log.Warn("http request returned client error", fields...)
		}
	}
}

type Server struct {
	appointments *application.AppointmentService
	reminders    *applicationv2.AppointmentLifecycleService
	calendar     *applicationv2.CalendarService
	services     *application.ServiceService
	log          *zap.Logger
}

func NewServer(appointments *application.AppointmentService, reminders *applicationv2.AppointmentLifecycleService, calendar *applicationv2.CalendarService, services *application.ServiceService, log *zap.Logger) *Server {
	if log == nil {
		log = zap.NewNop()
	}
	return &Server{appointments: appointments, reminders: reminders, calendar: calendar, services: services, log: log}
}

var _ httpserver.StrictServerInterface = (*Server)(nil)
