package di

import "github.com/petretiandrea/beaesthetic-backend/appointment/internal/infra/messaging"

func (d *DiContainer) GetAppointmentLifecycleConsumer() *messaging.Consumer {
	return singleton(d, "appointmentLifecycleConsumer", func() *messaging.Consumer {
		return messaging.NewConsumer(
			d.Config.RabbitMQ.URL,
			d.Config.RabbitMQ.AppointmentInternalJobQueue,
			messaging.NewAppointmentLifecycleConsumer(d.GetAppointmentLifecycleServiceV2(), d.Log),
			d.Log,
		)
	})
}

func (d *DiContainer) GetNotificationOutcomeQueueConsumer() *messaging.Consumer {
	return singleton(d, "notificationOutcomeQueueConsumer", func() *messaging.Consumer {
		return messaging.NewConsumer(
			d.Config.RabbitMQ.URL,
			d.Config.RabbitMQ.CustomerNotificationOutcomesQueue,
			messaging.NewNotificationOutcomeQueueConsumer(d.GetAppointmentLifecycleServiceV2(), d.Log),
			d.Log,
		)
	})
}
