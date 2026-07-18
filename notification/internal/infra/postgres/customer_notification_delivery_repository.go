package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
	outboxpostgres "github.com/petretiandrea/outbox-go/pkg/outbox/postgres"
)

const channelCustomerNotificationsConfirmed = "customer_notifications.confirmed"

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
			id, idempotency_key, correlation_key, customer_id, notification_type, notification_channel,
			template_values, status, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (idempotency_key) DO NOTHING
	`, delivery.ID, delivery.IdempotencyKey, delivery.CorrelationKey, delivery.CustomerID, delivery.NotificationType, delivery.NotificationChannel, delivery.TemplateValues, delivery.Status, delivery.CreatedAt)
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

func (repo *CustomerNotificationRepository) MarkSentBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, sentAt time.Time) (bool, error) {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin customer notification sent transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var correlationKey string
	err = tx.QueryRow(ctx, `
		UPDATE customer_notifications
		SET status = 'sent', sent_at = $2
		WHERE id = (
			SELECT customer_notification_id
			FROM customer_notification_sms_gateway_messages
			WHERE sms_gateway_message_id = $1
		)
		RETURNING correlation_key
	`, smsGatewayMessageID, sentAt).Scan(&correlationKey)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("mark customer notification sent: %w", err)
	}
	if err := publishCustomerNotificationConfirmed(ctx, tx, correlationKey, sentAt); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit customer notification sent transaction: %w", err)
	}
	return true, nil
}

func (repo *CustomerNotificationRepository) MarkFailedBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, failedAt time.Time) (bool, error) {
	commandTag, err := repo.db.Exec(ctx, `
		UPDATE customer_notifications
		SET status = 'failed', failed_at = $2
		WHERE id = (
			SELECT customer_notification_id
			FROM customer_notification_sms_gateway_messages
			WHERE sms_gateway_message_id = $1
		)
	`, smsGatewayMessageID, failedAt)
	if err != nil {
		return false, fmt.Errorf("mark customer notification failed: %w", err)
	}
	return commandTag.RowsAffected() > 0, nil
}

func publishCustomerNotificationConfirmed(ctx context.Context, tx pgx.Tx, correlationKey string, occurredAt time.Time) error {
	payload, err := json.Marshal(map[string]string{"notificationId": correlationKey})
	if err != nil {
		return fmt.Errorf("marshal customer notification confirmed event: %w", err)
	}
	publisher, err := outboxpostgres.NewPublisher(tx, outboxpostgres.PublisherConfig{TableName: "outbox_messages"})
	if err != nil {
		return fmt.Errorf("create customer notification confirmed publisher: %w", err)
	}
	if err := publisher.Publish(ctx, outbox.Message{
		ID:          uuid.NewString(),
		Channel:     outbox.Channel(channelCustomerNotificationsConfirmed),
		AffinityKey: outbox.AffinityKey(correlationKey),
		Payload:     payload,
		Metadata:    outbox.Metadata{},
		OccurredAt:  occurredAt,
	}); err != nil {
		return fmt.Errorf("publish customer notification confirmed event: %w", err)
	}
	return nil
}
