package di

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/postgres"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/provider"
)

func (d *DiContainer) GetNotificationService() *application.NotificationService {
	return singleton(d, "notificationService", func() *application.NotificationService {
		return application.NewNotificationService(d.GetNotificationRepository(), d.GetNotificationProvider())
	})
}

func (d *DiContainer) GetNotificationProvider() application.NotificationProvider {
	return singleton(d, "notificationProvider", func() application.NotificationProvider {
		return provider.NewCompoundProvider(provider.NewSMSProvider(d.Config.SMSGateway))
	})
}

func (d *DiContainer) GetNotificationRepository() application.NotificationRepository {
	return singleton(d, "notificationRepository", func() application.NotificationRepository {
		return postgres.NewNotificationRepository(d.GetPostgresDatabase())
	})
}

func (d *DiContainer) GetPostgresDatabase() *pgxpool.Pool {
	// pgxpool
	return singletonWithError(d, "postgresDatabase", func() (*pgxpool.Pool, error) {
		return pgxpool.New(context.Background(), d.Config.Postgres.DSN)
	})
}
