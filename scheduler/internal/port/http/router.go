package http

import (
	"net/http"
	"time"

	"github.com/beaesthetic/scheduler/internal/application"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

type Router struct {
	engine *gin.Engine
}

type createScheduleRequest struct {
	ScheduleAt time.Time      `json:"scheduleAt" binding:"required"`
	Route      string         `json:"route" binding:"required"`
	Data       map[string]any `json:"data" binding:"required"`
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
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	r.engine.GET("/actuator/health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	r.engine.GET("/actuator/health/readiness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	r.engine.PUT("/schedules/:scheduleId", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("scheduleId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scheduleId"})
			return
		}

		var request createScheduleRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := service.Schedule(c.Request.Context(), id, request.ScheduleAt, request.Route, request.Data); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to schedule job"})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"scheduleId": id})
	})

	r.engine.DELETE("/schedules/:scheduleId", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("scheduleId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scheduleId"})
			return
		}

		if err := service.Delete(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete schedule job"})
			return
		}

		c.Status(http.StatusNoContent)
	})
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
