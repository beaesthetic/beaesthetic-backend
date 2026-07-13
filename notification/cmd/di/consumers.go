package di

import "github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/messaging"

func (d *DiContainer) GetNotificationOutboxConsumer() *messaging.Consumer {
	return singleton(d, "notificationOutboxConsumer", func() *messaging.Consumer {
		return messaging.NewConsumer(
			d.Config.RabbitMQ.URL,
			d.Config.RabbitMQ.NotificationQueue,
			messaging.NewNotificationOutboxConsumer(d.GetNotificationService()),
			d.Log,
		)
	})
}

func (d *DiContainer) GetCustomerNotificationConsumer() *messaging.Consumer {
	return singleton(d, "customerNotificationConsumer", func() *messaging.Consumer {
		return messaging.NewConsumer(
			d.Config.RabbitMQ.URL,
			d.Config.RabbitMQ.CustomerNotificationQueue,
			messaging.NewCustomerNotificationConsumer(d.GetCustomerNotificationService(), d.Log),
			d.Log,
		)
	})
}
