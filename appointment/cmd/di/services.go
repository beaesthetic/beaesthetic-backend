package di

import (
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/postgres"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/client/customer"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/client/notification"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/client/scheduler"
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

func (d *DiContainer) GetAppointmentLifecycleHandler() *application.AppointmentLifecycleHandler {
	return singleton(d, "appointmentLifecycleHandler", func() *application.AppointmentLifecycleHandler {
		return application.NewAppointmentLifecycleHandler(
			d.GetAppointmentService(),
			d.GetCustomerRegistry(),
			d.GetReminderScheduler(),
			d.GetNotificationClient(),
			d.GetClock(),
			d.Config.Reminder.NoSendThreshold,
			d.Config.Reminder.ImmediateSendThreshold,
		)
	})
}

func (d *DiContainer) GetClock() application.Clock {
	return singleton(d, "clock", func() application.Clock {
		return application.SystemClock{}
	})
}

func (d *DiContainer) GetCustomerRegistry() application.CustomerRegistry {
	return singletonWithError(d, "customerRegistry", func() (application.CustomerRegistry, error) {
		return customer.NewCustomerRegistry(d.Config.Remote.CustomerURL)
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

func (d *DiContainer) GetNotificationClient() *notification.NotificationClient {
	return singletonWithError(d, "notificationClient", func() (*notification.NotificationClient, error) {
		return notification.NewNotificationClient(d.Config.Remote.NotificationURL)
	})
}

func (d *DiContainer) GetSchedulerClient() *scheduler.SchedulerClient {
	return singletonWithError(d, "schedulerClient", func() (*scheduler.SchedulerClient, error) {
		return scheduler.NewSchedulerClient(d.Config.Remote.SchedulerURL)
	})
}

func (d *DiContainer) GetReminderScheduler() application.ReminderScheduler {
	return singleton(d, "reminderScheduler", func() application.ReminderScheduler {
		return scheduler.NewReminderScheduler(d.GetSchedulerClient(), d.Config.RabbitMQ.SchedulerQueue)
	})
}
