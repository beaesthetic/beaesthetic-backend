package di

import (
	nethttp "net/http"

	"github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server"
)

func (d *DiContainer) GetHttpServer() *nethttp.Server {
	return singleton(d, "httpServer", func() *nethttp.Server {
		ginEngine := server.New(d.GetHttpHandlers(), d.Log)
		return &nethttp.Server{Addr: d.Config.HTTP.Addr, Handler: ginEngine}
	})
}

func (d *DiContainer) GetHttpHandlers() *server.HttpHandlers {
	return singleton(d, "httpHandlers", func() *server.HttpHandlers {
		return &server.HttpHandlers{
			Customer: d.CustomerHttpHandler(),
			Fidelity: d.CustomerHttpHandler(),
			Wallet:   d.CustomerHttpHandler(),
			DB:       d.GetPostgresDatabase(),
		}
	})
}

func (d *DiContainer) CustomerHttpHandler() *server.Server {
	return singleton(d, "customerHttpHandler", func() *server.Server {
		return server.NewServer(d.GetCustomerService(), d.GetFidelityService(), d.GetWalletService(), d.Log)
	})
}
