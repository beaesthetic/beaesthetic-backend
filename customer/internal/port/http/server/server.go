package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/application"
	cacheinfra "github.com/petretiandrea/beaesthetic-backend/customer/internal/infra/cache"
	customerapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/customer"
	fidelityapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/fidelity"
	walletapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/wallet"
	"go.uber.org/zap"
)

type HttpHandlers struct {
	Customer customerapi.StrictServerInterface
	Fidelity fidelityapi.StrictServerInterface
	Wallet   walletapi.StrictServerInterface
	DB       *sql.DB
}

func New(handlers *HttpHandlers, log *zap.Logger) *gin.Engine {
	if log == nil {
		log = zap.NewNop()
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginErrorLogger(log))

	customerapi.RegisterHandlers(r, customerapi.NewStrictHandlerWithOptions(handlers.Customer, nil, customerapi.StrictGinServerOptions{
		RequestErrorHandlerFunc:  requestErrorHandler(log),
		HandlerErrorFunc:         handlerErrorHandler(log),
		ResponseErrorHandlerFunc: responseErrorHandler(log),
	}))
	fidelityapi.RegisterHandlers(r, fidelityapi.NewStrictHandlerWithOptions(handlers.Fidelity, nil, fidelityapi.StrictGinServerOptions{
		RequestErrorHandlerFunc:  requestErrorHandler(log),
		HandlerErrorFunc:         handlerErrorHandler(log),
		ResponseErrorHandlerFunc: responseErrorHandler(log),
	}))
	walletapi.RegisterHandlers(r, walletapi.NewStrictHandlerWithOptions(handlers.Wallet, nil, walletapi.StrictGinServerOptions{
		RequestErrorHandlerFunc:  requestErrorHandler(log),
		HandlerErrorFunc:         handlerErrorHandler(log),
		ResponseErrorHandlerFunc: responseErrorHandler(log),
	}))
	r.GET("/health", healthCheck(handlers.DB))
	return r
}

func healthCheck(db *sql.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if db == nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
			return
		}
		if err := db.PingContext(ctx.Request.Context()); err != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "up"})
	}
}

func requestErrorHandler(log *zap.Logger) func(*gin.Context, error) {
	return func(ctx *gin.Context, err error) {
		log.Warn("http request error", zap.Error(err), zap.String("method", ctx.Request.Method), zap.String("path", ctx.FullPath()))
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	}
}

func handlerErrorHandler(log *zap.Logger) func(*gin.Context, error) {
	return func(ctx *gin.Context, err error) {
		status := http.StatusBadRequest
		if isNotFound(err) {
			status = http.StatusNotFound
		}
		log.Warn("http handler error", zap.Error(err), zap.String("method", ctx.Request.Method), zap.String("path", ctx.FullPath()), zap.Int("status", status))
		ctx.JSON(status, gin.H{"message": err.Error()})
	}
}

func responseErrorHandler(log *zap.Logger) func(*gin.Context, error) {
	return func(ctx *gin.Context, err error) {
		log.Error("http response error", zap.Error(err), zap.String("method", ctx.Request.Method), zap.String("path", ctx.FullPath()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
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

func isNotFound(err error) bool {
	return errors.Is(err, application.ErrNotFound) || strings.Contains(strings.ToLower(err.Error()), "not found")
}

type CustomerCacheTTL struct {
	Customers       time.Duration
	CustomersSearch time.Duration
}

type Server struct {
	customers *application.CustomerService
	fidelity  *application.FidelityService
	wallet    *application.WalletService
	cache     *cacheinfra.Cache
	cacheTTL  CustomerCacheTTL
	log       *zap.Logger
}

func NewServer(customers *application.CustomerService, fidelity *application.FidelityService, wallet *application.WalletService, cache *cacheinfra.Cache, cacheTTL CustomerCacheTTL, log *zap.Logger) *Server {
	if log == nil {
		log = zap.NewNop()
	}
	return &Server{customers: customers, fidelity: fidelity, wallet: wallet, cache: cache, cacheTTL: cacheTTL, log: log}
}

var _ customerapi.StrictServerInterface = (*Server)(nil)
var _ fidelityapi.StrictServerInterface = (*Server)(nil)
var _ walletapi.StrictServerInterface = (*Server)(nil)
