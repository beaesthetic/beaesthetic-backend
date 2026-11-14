package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	applicationv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/application/v2"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/health"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
	"go.uber.org/zap"
)

type HttpHandlers struct {
	Calendar      *Server
	HealthChecker health.HealthCheckHandler
}

func New(handlers *HttpHandlers, log *zap.Logger, verifier *auth.Verifier) (*gin.Engine, error) {
	if verifier == nil {
		return nil, fmt.Errorf("appointment authentication requires a verifier")
	}
	return newRouter(handlers, log, func(permissions ...string) gin.HandlerFunc {
		return authorize(verifier, permissions...)
	}), nil
}

func newRouter(handlers *HttpHandlers, log *zap.Logger, security func(...string) gin.HandlerFunc) *gin.Engine {
	if log == nil {
		log = zap.NewNop()
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginErrorLogger(log))
	if security == nil {
		security = func(...string) gin.HandlerFunc { return func(*gin.Context) {} }
	}
	if handlers.Calendar != nil {
		registerCalendarProtoRoutes(r, handlers.Calendar, security)
	}
	r.GET("/health", handlers.HealthChecker)
	return r
}

func authorize(verifier *auth.Verifier, permissions ...string) gin.HandlerFunc {
	policy := auth.Policy{Permissions: permissions, Scope: auth.ScopeMembership}
	require := auth.Require(policy)
	return func(ctx *gin.Context) {
		accepted := false
		handler := verifier.Authenticate(require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx.Request = r
			accepted = true
		})))
		handler.ServeHTTP(ctx.Writer, ctx.Request)
		if !accepted {
			ctx.Abort()
		}
	}
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
	reminders *applicationv2.AppointmentLifecycleService
	calendar  *applicationv2.CalendarService
	services  *application.ServiceService
	log       *zap.Logger
}

func NewServer(reminders *applicationv2.AppointmentLifecycleService, calendar *applicationv2.CalendarService, services *application.ServiceService, log *zap.Logger) *Server {
	if log == nil {
		log = zap.NewNop()
	}
	return &Server{reminders: reminders, calendar: calendar, services: services, log: log}
}
