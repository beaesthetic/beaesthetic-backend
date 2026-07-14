package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
)

type CustomerNotificationRepository struct {
	db *pgxpool.Pool
}

func NewCustomerNotificationRepository(db *pgxpool.Pool) *CustomerNotificationRepository {
	return &CustomerNotificationRepository{db: db}
}

func (repo *CustomerNotificationRepository) Exists(ctx context.Context, idempotencyKey string) (bool, error) {
	var exists bool
	if err := repo.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM customer_notifications WHERE idempotency_key = $1
		)
	`, idempotencyKey).Scan(&exists); err != nil {
		return false, fmt.Errorf("check customer notification idempotency: %w", err)
	}
	return exists, nil
}

func (repo *CustomerNotificationRepository) CreatePending(ctx context.Context, delivery application.CustomerNotificationRecord) (bool, error) {
	commandTag, err := repo.db.Exec(ctx, `
		INSERT INTO customer_notifications (
			id, idempotency_key, customer_id, notification_type, notification_channel,
			template_values, status, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (idempotency_key) DO NOTHING
	`, delivery.ID, delivery.IdempotencyKey, delivery.CustomerID, delivery.NotificationType, delivery.NotificationChannel, delivery.TemplateValues, delivery.Status, delivery.CreatedAt)
	if err != nil {
		return false, fmt.Errorf("create pending customer notification: %w", err)
	}
	return commandTag.RowsAffected() > 0, nil
}

func (repo *CustomerNotificationRepository) SaveSMSGatewayDispatch(ctx context.Context, dispatch application.SMSGatewayDispatch) error {
	_, err := repo.db.Exec(ctx, `
		INSERT INTO customer_notification_sms_gateway_messages (
			id, customer_notification_id, sms_gateway_message_id, created_at
		)
		VALUES ($1, $2, $3, $4)
	`, dispatch.ID, dispatch.CustomerNotificationID, dispatch.SMSGatewayMessageID, dispatch.CreatedAt)
	if err != nil {
		return fmt.Errorf("save customer notification sms gateway message: %w", err)
	}
	return nil
}

func (repo *CustomerNotificationRepository) MarkSentBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, sentAt time.Time) error {
	_, err := repo.db.Exec(ctx, `
		UPDATE customer_notifications
		SET status = 'sent', sent_at = $2
		WHERE id = (
			SELECT customer_notification_id
			FROM customer_notification_sms_gateway_messages
			WHERE sms_gateway_message_id = $1
		)
	`, smsGatewayMessageID, sentAt)
	if err != nil {
		return fmt.Errorf("mark customer notification sent: %w", err)
	}
	return nil
}

func (repo *CustomerNotificationRepository) MarkFailedBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, failedAt time.Time) error {
	_, err := repo.db.Exec(ctx, `
		UPDATE customer_notifications
		SET status = 'failed', failed_at = $2
		WHERE id = (
			SELECT customer_notification_id
			FROM customer_notification_sms_gateway_messages
			WHERE sms_gateway_message_id = $1
		)
	`, smsGatewayMessageID, failedAt)
	if err != nil {
		return fmt.Errorf("mark customer notification failed: %w", err)
	}
	return nil
}
