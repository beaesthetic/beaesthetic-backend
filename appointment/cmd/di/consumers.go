package di

import "github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/messaging"

func (d *DiContainer) GetSchedulerQueueConsumer() *messaging.Consumer {
	return singleton(d, "schedulerQueueConsumer", func() *messaging.Consumer {
		return messaging.NewConsumer(
			d.Config.RabbitMQ.URL,
			d.Config.RabbitMQ.SchedulerQueue,
			messaging.NewSchedulerQueueConsumer(d.GetReminderSender(), d.Log),
			d.Log,
		)
	})
}

func (d *DiContainer) GetAppointmentLifecycleConsumer() *messaging.Consumer {
	return singleton(d, "appointmentLifecycleConsumer", func() *messaging.Consumer {
		return messaging.NewConsumer(
			d.Config.RabbitMQ.URL,
			d.Config.RabbitMQ.AppointmentInternalJobQueue,
			messaging.NewAppointmentLifecycleConsumer(d.GetAppointmentLifecycleHandler(), d.Log),
			d.Log,
		)
	})
}

func (d *DiContainer) GetNotificationConfirmQueueConsumer() *messaging.Consumer {
	return singleton(d, "notificationConfirmQueueConsumer", func() *messaging.Consumer {
		return messaging.NewConsumer(
			d.Config.RabbitMQ.URL,
			d.Config.RabbitMQ.NotificationConfirmQueue,
			messaging.NewNotificationConfirmQueueConsumer(d.GetAppointmentService(), d.Log),
			d.Log,
		)
	})
}
