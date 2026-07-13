package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
)

type CustomerNotificationDeliveryRepository struct {
	db *pgxpool.Pool
}

func NewCustomerNotificationDeliveryRepository(db *pgxpool.Pool) *CustomerNotificationDeliveryRepository {
	return &CustomerNotificationDeliveryRepository{db: db}
}

func (repo *CustomerNotificationDeliveryRepository) Exists(ctx context.Context, idempotencyKey string) (bool, error) {
	var exists bool
	if err := repo.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM customer_notification_deliveries WHERE idempotency_key = $1
		)
	`, idempotencyKey).Scan(&exists); err != nil {
		return false, fmt.Errorf("check customer notification idempotency: %w", err)
	}
	return exists, nil
}

func (repo *CustomerNotificationDeliveryRepository) Save(ctx context.Context, delivery application.CustomerNotificationDelivery) error {
	_, err := repo.db.Exec(ctx, `
		INSERT INTO customer_notification_deliveries (
			idempotency_key, notification_id, customer_id, notification_type, notification_channel, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (idempotency_key) DO NOTHING
	`, delivery.IdempotencyKey, delivery.NotificationID, delivery.CustomerID, delivery.NotificationType, delivery.NotificationChannel, delivery.CreatedAt)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("save customer notification delivery: %w", err)
	}
	return nil
}
