package di

import (
	nethttp "net/http"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/health"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/server"
)

func (d *DiContainer) GetHttpServer() *nethttp.Server {
	return singleton(d, "httpServer", func() *nethttp.Server {
		ginEngine := server.New(d.GetHttpHandlers())
		return &nethttp.Server{Addr: d.Config.HTTP.Addr, Handler: ginEngine}
	})
}

func (d *DiContainer) GetHttpHandlers() *server.HttpHandlers {
	return singleton(d, "httpHandlers", func() *server.HttpHandlers {
		return &server.HttpHandlers{
			Appointment:   d.AppointmentHttpHandler(),
			HealthChecker: d.HealthCheckHandler(),
		}
	})
}

func (d *DiContainer) AppointmentHttpHandler() *server.Server {
	return singleton(d, "appointmentHttpHandler", func() *server.Server {
		return server.NewServer(d.GetAppointmentService(), d.Log)
	})
}

func (d *DiContainer) HealthCheckHandler() health.HealthCheckHandler {
	return singleton(d, "healthCheckHandler", func() health.HealthCheckHandler {
		return health.NewHealthCheckHandler(d.GetPostgresDatabase())
	})
}
