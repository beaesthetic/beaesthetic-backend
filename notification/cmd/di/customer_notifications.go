package di

import (
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	infracustomer "github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/customer"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/postgres"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/provider"
	notificationtemplate "github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/template"
)

func (d *DiContainer) GetCustomerNotificationService() *application.CustomerNotificationService {
	return singleton(d, "customerNotificationService", func() *application.CustomerNotificationService {
		return application.NewCustomerNotificationService(
			d.GetCustomerClient(),
			d.GetCustomerNotificationTemplateRenderer(),
			d.GetCustomerNotificationRepository(),
			d.GetCustomerNotificationSMSDispatcher(),
		)
	})
}

func (d *DiContainer) GetCustomerClient() application.CustomerReader {
	return singletonWithError(d, "customerClient", func() (application.CustomerReader, error) {
		return infracustomer.NewClient(d.Config.CustomerService.URL)
	})
}

func (d *DiContainer) GetCustomerNotificationTemplateRenderer() application.CustomerNotificationTemplateRenderer {
	return singleton(d, "customerNotificationTemplateRenderer", func() application.CustomerNotificationTemplateRenderer {
		return notificationtemplate.NewRenderer(d.Config.Templates.Path)
	})
}

func (d *DiContainer) GetCustomerNotificationRepository() application.CustomerNotificationRepository {
	return singleton(d, "customerNotificationIdempotencyRepository", func() application.CustomerNotificationRepository {
		return postgres.NewCustomerNotificationRepository(d.GetPostgresContextDB(), d.GetOutboxPublisher())
	})
}

func (d *DiContainer) GetCustomerNotificationSMSDispatcher() application.CustomerNotificationSMSDispatcher {
	return singleton(d, "customerNotificationSMSDispatcher", func() application.CustomerNotificationSMSDispatcher {
		return provider.NewCustomerNotificationSMSDispatcher(d.Config.SMSGateway)
	})
}
