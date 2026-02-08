package http

import (
	"github.com/beaesthetic/consent-service/internal/application"
	"github.com/beaesthetic/consent-service/internal/port/http/api"
	"github.com/gin-gonic/gin"
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
	engine.Use(gin.Logger())

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

// corsMiddleware returns a CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
