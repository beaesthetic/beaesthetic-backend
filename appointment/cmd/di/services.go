package di

import (
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/httpclient"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/postgres"
)

func (d *DiContainer) GetAppointmentService() *application.AppointmentService {
	return singleton(d, "appointmentService", func() *application.AppointmentService {
		return application.NewAppointmentService(
			d.GetAppointmentRepository(),
			d.GetCustomerRegistry(),
			d.Config.Reminder.TriggerBefore,
			d.GetClock(),
		)
	})
}

func (d *DiContainer) GetServiceService() *application.ServiceService {
	return singleton(d, "serviceService", func() *application.ServiceService {
		return application.NewServiceService(d.GetServiceRepository())
	})
}

func (d *DiContainer) GetClock() application.Clock {
	return singleton(d, "clock", func() application.Clock {
		return application.SystemClock{}
	})
}

func (d *DiContainer) GetCustomerRegistry() application.CustomerRegistry {
	return singletonWithError(d, "customerRegistry", func() (application.CustomerRegistry, error) {
		return httpclient.NewCustomerRegistry(d.Config.Remote.CustomerURL)
	})
}

func (d *DiContainer) GetAppointmentRepository() application.AppointmentRepository {
	return singleton(d, "appointmentRepository", func() application.AppointmentRepository {
		return d.GetPostgresRepository()
	})
}

func (d *DiContainer) GetServiceRepository() application.ServiceRepository {
	return singleton(d, "serviceRepository", func() application.ServiceRepository {
		return d.GetPostgresRepository()
	})
}

func (d *DiContainer) GetPostgresRepository() *postgres.Repository {
	return singleton(d, "postgresRepository", func() *postgres.Repository {
		return postgres.NewRepository(d.GetPostgresContextDB(), d.GetOutboxPublisher())
	})
}
