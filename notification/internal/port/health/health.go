package health

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/gin-gonic/gin"
)

type HealthCheckHandler = func(ctx *gin.Context)

type Pinger interface {
	Ping(ctx context.Context) error
}

func NewHealthCheckHandler(pinger Pinger) func(ctx *gin.Context) {
	healthCheck := createHealthCheck(pinger)
	return func(ctx *gin.Context) {
		result := healthCheck.Check(ctx)
		if result.Status == health.StatusDown {
			ctx.JSONP(http.StatusInternalServerError, result)
			return
		}
		ctx.JSONP(http.StatusOK, result)
	}
}

func FilterHealthCheck(request *http.Request) bool {
	return !strings.Contains(request.URL.Path, "health")
}

func createHealthCheck(db Pinger) health.Checker {
	return health.NewChecker(
		health.WithCacheDuration(5*time.Minute),
		health.WithCheck(health.Check{
			Name:    "postgres",
			Timeout: 2 * time.Second,
			Check: func(ctx context.Context) error {
				return db.Ping(ctx)
			},
		}),
	)
}
