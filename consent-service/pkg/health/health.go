package health

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterGinHealthCheck(gin *gin.Engine, mongo *mongo.Client) {
	gin.GET("/health", healthCheckHandler(createHealthCheck(mongo)))
}

func FilterHealthCheck(request *http.Request) bool {
	return !strings.Contains(request.URL.Path, "health")
}

func healthCheckHandler(healthCheck health.Checker) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		result := healthCheck.Check(ctx)
		if result.Status == health.StatusDown {
			ctx.JSONP(http.StatusInternalServerError, result)
		} else {
			ctx.JSONP(http.StatusOK, result)
		}
	}
}

func createHealthCheck(mongo *mongo.Client) health.Checker {
	checker := health.NewChecker(
		health.WithCacheDuration(5*time.Minute),
		health.WithCheck(health.Check{
			Name:    "mongo",
			Timeout: 2 * time.Second,
			Check: func(ctx context.Context) error {
				return mongo.Ping(ctx, nil)
			},
		}),
	)

	return checker
}
