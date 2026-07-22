package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/postgres/queries"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

const channelCustomerNotificationsConfirmed = "customer_notifications.confirmed"

type CustomerNotificationRepository struct {
	db        *ContextDB
	queries   *queries.Queries
	publisher outbox.Publisher
}

func NewCustomerNotificationRepository(db *ContextDB, publisher outbox.Publisher) *CustomerNotificationRepository {
	return &CustomerNotificationRepository{db: db, queries: queries.New(db), publisher: publisher}
}

func (repo *CustomerNotificationRepository) Exists(ctx context.Context, idempotencyKey string) (bool, error) {
	exists, err := repo.queries.ExistsCustomerNotificationByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return false, fmt.Errorf("check customer notification idempotency: %w", err)
	}
	return exists, nil
}

func (repo *CustomerNotificationRepository) CreatePending(ctx context.Context, delivery application.CustomerNotificationRecord) (bool, error) {
	templateValues, err := json.Marshal(delivery.TemplateValues)
	if err != nil {
		return false, fmt.Errorf("marshal customer notification template values: %w", err)
	}
	rowsAffected, err := repo.queries.CreatePendingCustomerNotification(ctx, queries.CreatePendingCustomerNotificationParams{
		ID:                  delivery.ID,
		IdempotencyKey:      delivery.IdempotencyKey,
		CorrelationKey:      delivery.CorrelationKey,
		CustomerID:          delivery.CustomerID,
		NotificationType:    delivery.NotificationType,
		NotificationChannel: delivery.NotificationChannel,
		TemplateValues:      templateValues,
		Status:              delivery.Status,
		CreatedAt:           timestamptz(delivery.CreatedAt),
	})
	if err != nil {
		return false, fmt.Errorf("create pending customer notification: %w", err)
	}
	return rowsAffected > 0, nil
}

func (repo *CustomerNotificationRepository) SaveSMSGatewayDispatch(ctx context.Context, dispatch application.SMSGatewayDispatch) error {
	if err := repo.queries.SaveSMSGatewayDispatch(ctx, queries.SaveSMSGatewayDispatchParams{
		ID:                     dispatch.ID,
		CustomerNotificationID: dispatch.CustomerNotificationID,
		SmsGatewayMessageID:    dispatch.SMSGatewayMessageID,
		CreatedAt:              timestamptz(dispatch.CreatedAt),
	}); err != nil {
		return fmt.Errorf("save customer notification sms gateway message: %w", err)
	}
	return nil
}

func (repo *CustomerNotificationRepository) MarkSentBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, sentAt time.Time) (bool, error) {
	marked := false
	if err := repo.db.Tx(ctx, func(ctx context.Context) error {
		correlationKey, err := repo.queries.MarkCustomerNotificationSentBySMSGatewayMessageID(ctx, queries.MarkCustomerNotificationSentBySMSGatewayMessageIDParams{
			SmsGatewayMessageID: smsGatewayMessageID,
			SentAt:              timestamptz(sentAt),
		})
		if err == pgx.ErrNoRows {
			return nil
		}
		if err != nil {
			return fmt.Errorf("mark customer notification sent: %w", err)
		}
		if err := repo.publishCustomerNotificationConfirmed(ctx, correlationKey, sentAt); err != nil {
			return err
		}
		marked = true
		return nil
	}); err != nil {
		return false, fmt.Errorf("customer notification sent transaction: %w", err)
	}
	return marked, nil
}

func (repo *CustomerNotificationRepository) MarkFailedBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, failedAt time.Time) (bool, error) {
	rowsAffected, err := repo.queries.MarkCustomerNotificationFailedBySMSGatewayMessageID(ctx, queries.MarkCustomerNotificationFailedBySMSGatewayMessageIDParams{
		SmsGatewayMessageID: smsGatewayMessageID,
		FailedAt:            timestamptz(failedAt),
	})
	if err != nil {
		return false, fmt.Errorf("mark customer notification failed: %w", err)
	}
	return rowsAffected > 0, nil
}

func (repo *CustomerNotificationRepository) publishCustomerNotificationConfirmed(ctx context.Context, correlationKey string, occurredAt time.Time) error {
	payload, err := json.Marshal(map[string]string{"notificationId": correlationKey})
	if err != nil {
		return fmt.Errorf("marshal customer notification confirmed event: %w", err)
	}
	if err := repo.publisher.Publish(ctx, outbox.Message{
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

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}
