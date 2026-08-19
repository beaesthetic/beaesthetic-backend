package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	notificationcontracts "github.com/petretiandrea/beaesthetic-backend/core-contracts/notification"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/postgres/queries"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
	"google.golang.org/protobuf/encoding/protojson"
)

const channelCustomerNotificationsOutcome = "customer.notifications.outcomes"

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

func (repo *CustomerNotificationRepository) MarkDispatched(ctx context.Context, notificationID string, dispatchedAt time.Time) error {
	if _, err := repo.queries.MarkCustomerNotificationDispatched(ctx, queries.MarkCustomerNotificationDispatchedParams{
		ID:           notificationID,
		DispatchedAt: timestamptz(dispatchedAt),
	}); err != nil {
		return fmt.Errorf("mark customer notification dispatched: %w", err)
	}
	return nil
}

func (repo *CustomerNotificationRepository) MarkFailed(ctx context.Context, notificationID string, reason string, message string, failedAt time.Time) (bool, error) {
	return repo.markTerminal(ctx, application.CustomerNotificationStatusFailed, reason, message, failedAt, func(ctx context.Context) (customerNotificationOutcomeIdentity, error) {
		row, err := repo.queries.MarkCustomerNotificationFailed(ctx, queries.MarkCustomerNotificationFailedParams{
			ID:             notificationID,
			FailedAt:       timestamptz(failedAt),
			FailureReason:  text(reason),
			FailureMessage: text(message),
		})
		return customerNotificationOutcomeIdentityFromFailedRow(row), err
	})
}

func (repo *CustomerNotificationRepository) MarkSentBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, sentAt time.Time) (bool, error) {
	return repo.markTerminal(ctx, application.CustomerNotificationStatusSent, "", "", sentAt, func(ctx context.Context) (customerNotificationOutcomeIdentity, error) {
		row, err := repo.queries.MarkCustomerNotificationSentBySMSGatewayMessageID(ctx, queries.MarkCustomerNotificationSentBySMSGatewayMessageIDParams{
			SmsGatewayMessageID: smsGatewayMessageID,
			SentAt:              timestamptz(sentAt),
		})
		return customerNotificationOutcomeIdentityFromSentRow(row), err
	})
}

func (repo *CustomerNotificationRepository) MarkFailedBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, reason string, message string, failedAt time.Time) (bool, error) {
	return repo.markTerminal(ctx, application.CustomerNotificationStatusFailed, reason, message, failedAt, func(ctx context.Context) (customerNotificationOutcomeIdentity, error) {
		row, err := repo.queries.MarkCustomerNotificationFailedBySMSGatewayMessageID(ctx, queries.MarkCustomerNotificationFailedBySMSGatewayMessageIDParams{
			SmsGatewayMessageID: smsGatewayMessageID,
			FailedAt:            timestamptz(failedAt),
			FailureReason:       text(reason),
			FailureMessage:      text(message),
		})
		return customerNotificationOutcomeIdentityFromFailedByMessageRow(row), err
	})
}

func (repo *CustomerNotificationRepository) markTerminal(ctx context.Context, status string, reason string, message string, occurredAt time.Time, mark func(context.Context) (customerNotificationOutcomeIdentity, error)) (bool, error) {
	marked := false
	if err := repo.db.Tx(ctx, func(ctx context.Context) error {
		identity, err := mark(ctx)
		if err == pgx.ErrNoRows {
			return nil
		}
		if err != nil {
			return fmt.Errorf("mark customer notification %s: %w", status, err)
		}
		if err := repo.publishCustomerNotificationOutcome(ctx, identity, status, reason, message, occurredAt); err != nil {
			return err
		}
		marked = true
		return nil
	}); err != nil {
		return false, fmt.Errorf("customer notification %s transaction: %w", status, err)
	}
	return marked, nil
}

type customerNotificationOutcomeIdentity struct {
	CorrelationKey string
	IdempotencyKey string
	CustomerID     string
}

func (repo *CustomerNotificationRepository) publishCustomerNotificationOutcome(ctx context.Context, identity customerNotificationOutcomeIdentity, status string, reason string, message string, occurredAt time.Time) error {
	payload, err := protojson.Marshal(&notificationcontracts.CustomerNotificationOutcome{
		NotificationId: identity.CorrelationKey,
		Status:         customerNotificationOutcomeStatus(status),
		Reason:         reason,
		Message:        message,
		IdempotencyKey: identity.IdempotencyKey,
		CustomerId:     identity.CustomerID,
	})
	if err != nil {
		return fmt.Errorf("marshal customer notification outcome event: %w", err)
	}
	if err := repo.publisher.Publish(ctx, outbox.Message{
		ID:          uuid.NewString(),
		Channel:     outbox.Channel(channelCustomerNotificationsOutcome),
		AffinityKey: outbox.AffinityKey(identity.CorrelationKey),
		Payload:     payload,
		Metadata:    outbox.Metadata{},
		OccurredAt:  occurredAt,
	}); err != nil {
		return fmt.Errorf("publish customer notification outcome event: %w", err)
	}
	return nil
}

func customerNotificationOutcomeIdentityFromFailedRow(row queries.MarkCustomerNotificationFailedRow) customerNotificationOutcomeIdentity {
	return customerNotificationOutcomeIdentity{CorrelationKey: row.CorrelationKey, IdempotencyKey: row.IdempotencyKey, CustomerID: row.CustomerID}
}

func customerNotificationOutcomeIdentityFromSentRow(row queries.MarkCustomerNotificationSentBySMSGatewayMessageIDRow) customerNotificationOutcomeIdentity {
	return customerNotificationOutcomeIdentity{CorrelationKey: row.CorrelationKey, IdempotencyKey: row.IdempotencyKey, CustomerID: row.CustomerID}
}

func customerNotificationOutcomeIdentityFromFailedByMessageRow(row queries.MarkCustomerNotificationFailedBySMSGatewayMessageIDRow) customerNotificationOutcomeIdentity {
	return customerNotificationOutcomeIdentity{CorrelationKey: row.CorrelationKey, IdempotencyKey: row.IdempotencyKey, CustomerID: row.CustomerID}
}

func customerNotificationOutcomeStatus(status string) notificationcontracts.CustomerNotificationOutcomeStatus {
	switch status {
	case application.CustomerNotificationStatusSent:
		return notificationcontracts.CustomerNotificationOutcomeStatus_CUSTOMER_NOTIFICATION_OUTCOME_STATUS_SENT
	case application.CustomerNotificationStatusFailed:
		return notificationcontracts.CustomerNotificationOutcomeStatus_CUSTOMER_NOTIFICATION_OUTCOME_STATUS_FAILED
	default:
		return notificationcontracts.CustomerNotificationOutcomeStatus_CUSTOMER_NOTIFICATION_OUTCOME_STATUS_UNSPECIFIED
	}
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func text(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}
