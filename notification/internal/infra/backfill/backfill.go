package backfill

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg config.Config, log *zap.Logger) error {
	postgresDB, err := pgxpool.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer postgresDB.Close()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}
	defer mongoClient.Disconnect(ctx)

	collection := mongoClient.Database(cfg.Mongo.Database).Collection(cfg.Mongo.Collection)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("read legacy notifications: %w", err)
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var legacy legacyNotification
		if err := cursor.Decode(&legacy); err != nil {
			return fmt.Errorf("decode legacy notification: %w", err)
		}
		if err := upsert(ctx, postgresDB, legacy); err != nil {
			return err
		}
		count++
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	log.Info("backfill completed", zap.Int("count", count))
	return nil
}

type legacyNotification struct {
	ID              string          `bson:"_id"`
	Title           string          `bson:"title"`
	Content         string          `bson:"content"`
	IsSent          bool            `bson:"isSent"`
	IsSentConfirmed bool            `bson:"isSentConfirmed"`
	Channel         bson.Raw        `bson:"channel"`
	ChannelData     *legacyMetadata `bson:"channelData"`
	CreatedAt       time.Time       `bson:"createdAt"`
	UpdatedAt       time.Time       `bson:"updatedAt"`
}

type legacyMetadata struct {
	ProviderResourceID string `bson:"providerResourceId"`
}

func upsert(ctx context.Context, db *pgxpool.Pool, legacy legacyNotification) error {
	channelType, address, err := parseLegacyChannel(legacy.Channel)
	if err != nil {
		return err
	}
	var providerResourceID *string
	if legacy.ChannelData != nil && legacy.ChannelData.ProviderResourceID != "" {
		providerResourceID = &legacy.ChannelData.ProviderResourceID
	}
	_, err = db.Exec(ctx, `
		INSERT INTO notifications (
			id, title, content, is_sent, is_sent_confirmed, channel_type, channel_address,
			provider_resource_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO NOTHING
	`, legacy.ID, legacy.Title, legacy.Content, legacy.IsSent, legacy.IsSentConfirmed, channelType, address, providerResourceID, legacy.CreatedAt, legacy.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert backfilled notification %s: %w", legacy.ID, err)
	}
	return nil
}

func parseLegacyChannel(raw bson.Raw) (string, string, error) {
	var sms struct {
		Phone string `bson:"phone"`
	}
	if err := bson.Unmarshal(raw, &sms); err == nil && sms.Phone != "" {
		return "sms", sms.Phone, nil
	}
	var email struct {
		Email string `bson:"email"`
	}
	if err := bson.Unmarshal(raw, &email); err == nil && email.Email != "" {
		return "email", email.Email, nil
	}
	return "", "", fmt.Errorf("unsupported legacy channel %s", string(raw))
}
