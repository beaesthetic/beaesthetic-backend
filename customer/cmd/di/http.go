package di

import (
	nethttp "net/http"

	"github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server"
)

func (d *DiContainer) GetHttpServer() (*nethttp.Server, error) {
	verifier, err := d.Config.Auth.NewVerifier()
	if err != nil {
		return nil, err
	}
	router, err := server.New(d.GetHttpHandlers(), d.Log, verifier, d.Config.Auth.OrganizationID)
	if err != nil {
		return nil, err
	}
	return &nethttp.Server{Addr: d.Config.HTTP.Addr, Handler: router}, nil
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
		return server.NewServer(d.GetCustomerService(), d.GetFidelityService(), d.GetWalletService(), d.GetCustomerCache(), server.CustomerCacheTTL{Customers: d.Config.Redis.CustomersTTL, CustomersSearch: d.Config.Redis.CustomersSearchTTL}, d.Log)
	})
}
