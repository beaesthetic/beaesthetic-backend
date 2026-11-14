package di

import (
	nethttp "net/http"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/health"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/server"
)

func (d *DiContainer) GetHttpServer() (*nethttp.Server, error) {
	verifier, err := d.GetAuthVerifier()
	if err != nil {
		return nil, err
	}
	router, err := server.New(d.GetHttpHandlers(), d.Log, verifier)
	if err != nil {
		return nil, err
	}
	return &nethttp.Server{Addr: d.Config.HTTP.Addr, Handler: router}, nil
}

func (d *DiContainer) GetHttpHandlers() *server.HttpHandlers {
	return singleton(d, "httpHandlers", func() *server.HttpHandlers {
		return &server.HttpHandlers{
			Calendar:      d.CalendarHttpHandler(),
			HealthChecker: d.HealthCheckHandler(),
		}
	})
}

func (d *DiContainer) CalendarHttpHandler() *server.Server {
	return singleton(d, "calendarHttpHandler", func() *server.Server {
		return server.NewServer(d.GetAppointmentLifecycleServiceV2(), d.GetCalendarService(), d.GetServiceService(), d.Log)
	})
}

func (d *DiContainer) HealthCheckHandler() health.HealthCheckHandler {
	return singleton(d, "healthCheckHandler", func() health.HealthCheckHandler {
		return health.NewHealthCheckHandler(d.GetPostgresDatabase())
	})
}
