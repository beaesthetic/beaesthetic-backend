package di

import (
	"context"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/postgres"
)

func (d *DiContainer) context() context.Context { return context.Background() }

func (d *DiContainer) GetAppointmentService() *application.Service {
	return singleton(d, "appointmentService", func() *application.Service {
		return application.NewService(d.GetAppointmentRepository(), d.Config.Reminder.TriggerBefore)
	})
}

func (d *DiContainer) GetAppointmentRepository() application.Repository {
	return singleton(d, "appointmentRepository", func() application.Repository {
		return postgres.NewRepository(d.GetPostgresDatabase())
	})
}
