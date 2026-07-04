package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthCheckHandler = func(ctx *gin.Context)

type Pinger interface {
	Ping(ctx context.Context) error
}

func NewHealthCheckHandler(pinger Pinger) HealthCheckHandler {
	return func(ctx *gin.Context) {
		checkCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
		defer cancel()
		if err := pinger.Ping(checkCtx); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"status": "DOWN", "checks": []gin.H{{"name": "postgres", "status": "DOWN"}}})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "UP", "checks": []gin.H{{"name": "postgres", "status": "UP"}}})
	}
}
