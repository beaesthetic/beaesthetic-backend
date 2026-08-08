package di

import (
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	applicationv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/application/v2"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/jobs"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/messaging"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/postgres"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/client/customer"
)

type RiverReminderConfig struct {
	Queue       string
	Workers     int
	MaxAttempts int
}

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

func (d *DiContainer) GetCalendarService() *applicationv2.CalendarService {
	return singleton(d, "calendarService", func() *applicationv2.CalendarService {
		return applicationv2.NewCalendarService(
			d.GetPostgresRepository(),
			d.GetCustomerResolver(),
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
			d.GetAppointmentRepository(),
			d.GetReminderScheduler(),
			d.GetCustomerNotificationSender(),
			d.GetClock(),
			d.Config.Reminder.NoSendThreshold,
			d.Config.Reminder.ImmediateSendThreshold,
			d.Log,
		)
	})
}

func (d *DiContainer) GetCalendarLifecycleHandler() *applicationv2.CalendarLifecycleHandler {
	return singleton(d, "calendarLifecycleHandler", func() *applicationv2.CalendarLifecycleHandler {
		return applicationv2.NewCalendarLifecycleHandler(
			d.GetAppointmentLifecycleServiceV2(),
			d.GetAppointmentLifecycleHandler(),
		)
	})
}

func (d *DiContainer) GetAppointmentLifecycleServiceV2() *applicationv2.AppointmentLifecycleService {
	return singleton(d, "appointmentLifecycleServiceV2", func() *applicationv2.AppointmentLifecycleService {
		return applicationv2.NewAppointmentLifecycleService(
			d.GetPostgresRepository(),
			d.GetReminderScheduler(),
			d.GetCustomerNotificationSenderV2(),
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

func (d *DiContainer) GetCustomerResolver() applicationv2.CustomerResolver {
	return singletonWithError(d, "customerResolver", func() (applicationv2.CustomerResolver, error) {
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

func (d *DiContainer) GetCustomerNotificationSender() application.NotificationSender {
	return singleton(d, "customerNotificationSender", func() application.NotificationSender {
		return messaging.NewCustomerNotificationSender(d.GetOutboxPublisher())
	})
}

func (d *DiContainer) GetCustomerNotificationSenderV2() applicationv2.CalendarNotificationSender {
	return singleton(d, "customerNotificationSenderV2", func() applicationv2.CalendarNotificationSender {
		return messaging.NewCustomerNotificationSender(d.GetOutboxPublisher())
	})
}

func (d *DiContainer) GetReminderSender() *application.ReminderSender {
	return singleton(d, "reminderSender", func() *application.ReminderSender {
		return application.NewReminderSender(d.GetAppointmentService(), d.GetCustomerNotificationSender(), d.Log)
	})
}

func (d *DiContainer) GetReminderScheduler() *jobs.ReminderScheduler {
	return singleton(d, "reminderScheduler", func() *jobs.ReminderScheduler {
		riverConfig := d.GetRiverReminderConfig()
		return jobs.NewReminderScheduler(
			d.GetRiverJobInserter(),
			riverConfig.Queue,
			riverConfig.MaxAttempts,
			d.Log,
		)
	})
}

func (d *DiContainer) GetRiverReminderConfig() RiverReminderConfig {
	cfg := RiverReminderConfig{
		Queue:       d.Config.River.Queue,
		Workers:     d.Config.River.Workers,
		MaxAttempts: d.Config.River.MaxAttempts,
	}
	if cfg.Queue == "" {
		cfg.Queue = "appointment_reminders"
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 5
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	return cfg
}
