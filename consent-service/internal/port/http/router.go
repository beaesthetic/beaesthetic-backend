package http

import (
	"strings"
	"time"

	"github.com/beaesthetic/consent-service/internal/application"
	"github.com/beaesthetic/consent-service/internal/port/http/api"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Router configures and returns the HTTP router
type Router struct {
	engine *gin.Engine
}

// NewRouter creates a new Router with oapi-codegen generated handlers
func NewRouter(
	policyService *application.PolicyService,
	consentService *application.ConsentService,
	linkService *application.LinkService,
) *Router {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(requestLogger())

	// CORS middleware
	engine.Use(corsMiddleware())

	// Create the server implementation
	server := api.NewServer(policyService, consentService, linkService)

	// Create strict handler wrapper
	strictHandler := api.NewStrictHandler(server, nil)

	// Register all routes using the generated RegisterHandlers function
	api.RegisterHandlers(engine, strictHandler)

	return &Router{
		engine: engine,
	}
}

// Engine returns the underlying gin engine
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// requestLogger returns a middleware that logs HTTP requests (skips health checks)
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// Skip logging for health check endpoint
		if strings.Contains(c.Request.URL.Path, "/health") {
			return
		}

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		log.Info().
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("ip", clientIP).
			Msg("Request processed")
	}
}

// corsMiddleware returns a CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Tenant-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
