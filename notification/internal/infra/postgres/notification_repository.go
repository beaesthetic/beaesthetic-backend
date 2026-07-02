package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/domain"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
	outboxpostgres "github.com/petretiandrea/outbox-go/pkg/outbox/postgres"
)

const (
	channelNotifications          = "notifications"
	channelNotificationsConfirmed = "notifications.confirmed"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (repo *NotificationRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	var record notificationRecord
	err := repo.db.QueryRow(ctx, `
		SELECT id, title, content, is_sent, is_sent_confirmed, channel_type, channel_address,
		       provider_resource_id, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`, id).Scan(
		&record.ID,
		&record.Title,
		&record.Content,
		&record.IsSent,
		&record.IsSentConfirmed,
		&record.ChannelType,
		&record.ChannelAddress,
		&record.ProviderResourceID,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find notification: %w", err)
	}
	notification := record.toDomain()
	return &notification, nil
}

func (repo *NotificationRepository) Save(ctx context.Context, notification *domain.Notification) error {
	events := notification.PullEvents()
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin notification transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	record := newNotificationRecord(*notification)
	_, err = tx.Exec(ctx, `
		INSERT INTO notifications (
			id, title, content, is_sent, is_sent_confirmed, channel_type, channel_address,
			provider_resource_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			is_sent = EXCLUDED.is_sent,
			is_sent_confirmed = EXCLUDED.is_sent_confirmed,
			channel_type = EXCLUDED.channel_type,
			channel_address = EXCLUDED.channel_address,
			provider_resource_id = EXCLUDED.provider_resource_id,
			updated_at = EXCLUDED.updated_at
	`, record.ID, record.Title, record.Content, record.IsSent, record.IsSentConfirmed, record.ChannelType, record.ChannelAddress, record.ProviderResourceID, record.CreatedAt, record.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save notification: %w", err)
	}

	if len(events) > 0 {
		publisher, err := outboxpostgres.NewPublisher(tx, outboxpostgres.PublisherConfig{
			TableName: "outbox_messages",
		})
		if err != nil {
			return err
		}
		messages := make([]outbox.Message, 0, len(events))
		for _, event := range events {
			message, err := newOutboxMessage(event)
			if err != nil {
				return err
			}
			messages = append(messages, message)
		}
		if err := publisher.Publish(ctx, messages...); err != nil {
			return fmt.Errorf("publish outbox notification events: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit notification transaction: %w", err)
	}
	return nil
}

type notificationRecord struct {
	ID                 string
	Title              string
	Content            string
	IsSent             bool
	IsSentConfirmed    bool
	ChannelType        string
	ChannelAddress     string
	ProviderResourceID *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func newNotificationRecord(notification domain.Notification) notificationRecord {
	channelAddress := notification.Channel.Phone
	if notification.Channel.Type == domain.ChannelEmail {
		channelAddress = notification.Channel.Email
	}
	var providerResourceID *string
	if notification.ChannelMetadata != nil {
		providerResourceID = &notification.ChannelMetadata.ProviderResourceID
	}
	return notificationRecord{
		ID:                 notification.ID,
		Title:              notification.Title,
		Content:            notification.Content,
		IsSent:             notification.IsSent,
		IsSentConfirmed:    notification.IsSentConfirmed,
		ChannelType:        string(notification.Channel.Type),
		ChannelAddress:     channelAddress,
		ProviderResourceID: providerResourceID,
		CreatedAt:          notification.CreatedAt,
		UpdatedAt:          notification.UpdatedAt,
	}
}

func (record notificationRecord) toDomain() domain.Notification {
	channel := domain.Channel{Type: domain.ChannelType(record.ChannelType)}
	if channel.Type == domain.ChannelEmail {
		channel.Email = record.ChannelAddress
	} else {
		channel.Phone = record.ChannelAddress
	}
	var metadata *domain.ChannelMetadata
	if record.ProviderResourceID != nil {
		metadata = &domain.ChannelMetadata{ProviderResourceID: *record.ProviderResourceID}
	}
	return domain.HydrateNotification(record.ID, record.Title, record.Content, record.IsSent, record.IsSentConfirmed, channel, metadata, record.CreatedAt, record.UpdatedAt)
}

func newOutboxMessage(event domain.Event) (outbox.Message, error) {
	payload, err := json.Marshal(map[string]string{"notificationId": event.NotificationID()})
	if err != nil {
		return outbox.Message{}, fmt.Errorf("marshal notification event: %w", err)
	}
	channel := outbox.Channel(channelNotifications)
	if _, ok := event.(domain.NotificationSentConfirmed); ok {
		channel = outbox.Channel(channelNotificationsConfirmed)
	}
	return outbox.Message{
		ID:          uuid.NewString(),
		Channel:     channel,
		AffinityKey: outbox.AffinityKey(event.NotificationID()),
		Payload:     payload,
		Metadata:    outbox.Metadata{},
		OccurredAt:  time.Now().UTC(),
	}, nil
}
