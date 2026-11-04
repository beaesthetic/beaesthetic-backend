package server

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
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

func New(handlers *HttpHandlers, log *zap.Logger, verifier *auth.Verifier) (*gin.Engine, error) {
	if verifier == nil {
		return nil, fmt.Errorf("customer authentication requires a verifier")
	}
	return newRouter(handlers, log, func(permissions ...string) gin.HandlerFunc {
		return authorize(verifier, permissions...)
	}), nil
}

func newRouter(handlers *HttpHandlers, log *zap.Logger, security func(...string) gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	if log == nil {
		log = zap.NewNop()
	}
	r := gin.New()
	r.RedirectTrailingSlash = false
	r.Use(gin.Recovery())
	r.Use(ginErrorLogger(log))

	customerHandler := customerapi.NewStrictHandlerWithOptions(handlers.Customer, nil, customerapi.StrictGinServerOptions{
		RequestErrorHandlerFunc:  requestErrorHandler(log),
		HandlerErrorFunc:         handlerErrorHandler(log),
		ResponseErrorHandlerFunc: responseErrorHandler(log),
	})
	fidelityHandler := fidelityapi.NewStrictHandlerWithOptions(handlers.Fidelity, nil, fidelityapi.StrictGinServerOptions{
		RequestErrorHandlerFunc:  requestErrorHandler(log),
		HandlerErrorFunc:         handlerErrorHandler(log),
		ResponseErrorHandlerFunc: responseErrorHandler(log),
	})
	walletHandler := walletapi.NewStrictHandlerWithOptions(handlers.Wallet, nil, walletapi.StrictGinServerOptions{
		RequestErrorHandlerFunc:  requestErrorHandler(log),
		HandlerErrorFunc:         handlerErrorHandler(log),
		ResponseErrorHandlerFunc: responseErrorHandler(log),
	})

	// Permissions are declared alongside each route, including read-only POSTs.
	if security == nil {
		security = func(...string) gin.HandlerFunc { return func(*gin.Context) {} }
	}
	bindingError := func(ctx *gin.Context, err error, status int) {
		ctx.JSON(status, gin.H{"msg": err.Error()})
	}
	customers := customerapi.ServerInterfaceWrapper{Handler: customerHandler, ErrorHandler: bindingError}
	fidelity := fidelityapi.ServerInterfaceWrapper{Handler: fidelityHandler, ErrorHandler: bindingError}
	wallets := walletapi.ServerInterfaceWrapper{Handler: walletHandler, ErrorHandler: bindingError}

	r.GET("/admin/customers", security("customer.read"), customers.GetAllCustomers)
	r.POST("/admin/customers", security("customer.write"), customers.CreateCustomer)
	r.GET("/admin/customers/page/:direction", security("customer.read"), customers.GetCustomerByPage)
	r.DELETE("/admin/customers/:customerId", security("customer.write"), customers.DeleteCustomer)
	r.GET("/admin/customers/:customerId", security("customer.read"), customers.GetCustomerById)
	r.PUT("/admin/customers/:customerId", security("customer.write"), customers.UpdateCustomerById)
	r.GET("/customers/search", security("customer.read"), customers.SearchCustomer)
	r.POST("/customers/search:byPhone", security("customer.read"), customers.SearchCustomerByPhone)

	r.GET("/admin/fidelity-cards", security("customer.read"), fidelity.GetFidelityCards)
	r.POST("/admin/fidelity-cards", security("customer.write"), fidelity.CreateFidelityCard)
	r.GET("/admin/fidelity-cards/customers/:customerId", security("customer.read"), fidelity.GetFidelityCardsByCustomerId)
	r.DELETE("/admin/fidelity-cards/vouchers/:voucherId", security("customer.write"), fidelity.UseVoucher)
	r.GET("/admin/fidelity-cards/:cardId", security("customer.read"), fidelity.GetFidelityCardById)
	r.PUT("/admin/fidelity-cards/:cardId/purchase", security("customer.write"), fidelity.NotifyPurchase)

	r.GET("/admin/wallets", security("customer.read"), wallets.GetWallets)
	r.POST("/admin/wallets/giftCard", security("customer.write"), wallets.AddGiftCard)
	r.GET("/admin/wallets/:walletId", security("customer.read"), wallets.GetWalletById)
	r.PUT("/admin/wallets/:walletId/charge", security("customer.write"), wallets.ChargeWallet)

	r.GET("/admin/fidelity-cards/", security("customer.read"), fidelity.GetFidelityCards)
	r.POST("/admin/fidelity-cards/", security("customer.write"), fidelity.CreateFidelityCard)
	r.GET("/admin/wallets/", security("customer.read"), func(ctx *gin.Context) {
		walletHandler.GetWallets(ctx, walletapi.GetWalletsParams{Filter: stringPtrIfNotEmpty(ctx.Query("filter")), XOrganizationID: ctx.GetHeader(organizationHeader)})
	})
	r.POST("/admin/wallets/giftCard/", security("customer.write"), wallets.AddGiftCard)
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
