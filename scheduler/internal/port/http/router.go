package http

import (
	"net/http"
	"time"

	"github.com/beaesthetic/scheduler/api"
	"github.com/beaesthetic/scheduler/internal/application"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

type Router struct {
	engine *gin.Engine
}

type scheduleServer struct {
	service *application.SchedulerService
}

func NewRouter(service *application.SchedulerService, logger *zap.Logger) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(otelgin.Middleware("scheduler-service"))
	engine.Use(requestLogger(logger))

	router := &Router{engine: engine}
	router.register(service)
	return router
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func (r *Router) register(service *application.SchedulerService) {
	r.engine.GET("/health", healthHandler)
	r.engine.GET("/actuator/health/liveness", healthHandler)
	r.engine.GET("/actuator/health/readiness", healthHandler)

	api.RegisterHandlers(r.engine, &scheduleServer{service: service})
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP"})
}

func (s *scheduleServer) AddSchedule(c *gin.Context, scheduleId api.ScheduleId) {
	var request api.AddScheduleJSONRequestBody
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.service.Schedule(c.Request.Context(), scheduleId, request.ScheduleAt, request.Route, request.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to schedule job"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"scheduleId": scheduleId})
}

func (s *scheduleServer) RemoveSchedule(c *gin.Context, scheduleId api.ScheduleId) {
	if err := s.service.Delete(c.Request.Context(), scheduleId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete schedule job"})
		return
	}

	c.Status(http.StatusNoContent)
}

func requestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/actuator/health/liveness" || c.Request.URL.Path == "/actuator/health/readiness" {
			return
		}

		logger.Info("request processed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		)
	}
}
