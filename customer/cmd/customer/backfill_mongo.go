package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/customer/cmd/di"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type backfillMongoOptions struct {
	mongoURI      string
	mongoDatabase string
	batchSize     int32
}

type backfillStats struct {
	customers        int
	deletedCustomers int
	fidelityCards    int
	wallets          int
}

func backfillMongoCommand(envFile *string) *cobra.Command {
	opts := backfillMongoOptions{batchSize: 500}
	cmd := &cobra.Command{
		Use:   "backfill-mongo",
		Short: "Temporarily backfill customer data from MongoDB to PostgreSQL",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := di.NewDiContainer(cmd.Context(), *envFile)
			if err != nil {
				return err
			}
			defer func() { c.GetPostgresDatabase().Close(); _ = c.Log.Sync() }()

			if opts.mongoURI == "" {
				opts.mongoURI = c.Config.Mongo.URI
			}
			if opts.mongoDatabase == "" {
				opts.mongoDatabase = c.Config.Mongo.Database
			}
			if opts.mongoURI == "" || opts.mongoDatabase == "" {
				return fmt.Errorf("mongo uri and database must be set")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Minute)
			defer cancel()
			client, err := mongo.Connect(ctx, options.Client().ApplyURI(opts.mongoURI))
			if err != nil {
				return fmt.Errorf("connect mongo: %w", err)
			}
			defer func() { _ = client.Disconnect(context.Background()) }()
			if err := client.Ping(ctx, nil); err != nil {
				return fmt.Errorf("ping mongo: %w", err)
			}

			stats, err := backfillMongo(ctx, c.GetPostgresDatabase(), client.Database(opts.mongoDatabase), opts, c.Log)
			if err != nil {
				return err
			}
			c.Log.Info("mongo backfill completed",
				zap.Int("customers", stats.customers),
				zap.Int("deleted_customers", stats.deletedCustomers),
				zap.Int("fidelity_cards", stats.fidelityCards),
				zap.Int("wallets", stats.wallets),
			)
			fmt.Fprintf(cmd.OutOrStdout(), "customers=%d deleted_customers=%d fidelity_cards=%d wallets=%d\n", stats.customers, stats.deletedCustomers, stats.fidelityCards, stats.wallets)
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.mongoURI, "mongo-uri", "", "MongoDB URI; defaults to ENV_MONGO_URI")
	cmd.Flags().StringVar(&opts.mongoDatabase, "mongo-database", "", "MongoDB database; defaults to ENV_MONGO_DATABASE")
	cmd.Flags().Int32Var(&opts.batchSize, "batch-size", opts.batchSize, "Mongo cursor batch size")
	return cmd
}

func backfillMongo(ctx context.Context, db *sql.DB, mongoDB *mongo.Database, opts backfillMongoOptions, log *zap.Logger) (backfillStats, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return backfillStats{}, err
	}
	defer tx.Rollback()

	stats := backfillStats{}
	var count int
	count, err = backfillCustomers(ctx, tx, mongoDB.Collection("customers"), opts.batchSize, false)
	if err != nil {
		return stats, fmt.Errorf("backfill customers: %w", err)
	}
	stats.customers = count
	log.Info("backfilled customers", zap.Int("count", count))

	count, err = backfillCustomers(ctx, tx, mongoDB.Collection("delete_customers"), opts.batchSize, true)
	if err != nil {
		return stats, fmt.Errorf("backfill deleted customers: %w", err)
	}
	stats.deletedCustomers = count
	log.Info("backfilled deleted customers", zap.Int("count", count))

	count, err = backfillFidelityCards(ctx, tx, mongoDB.Collection("fidelitycards"), opts.batchSize)
	if err != nil {
		return stats, fmt.Errorf("backfill fidelity cards: %w", err)
	}
	stats.fidelityCards = count
	log.Info("backfilled fidelity cards", zap.Int("count", count))

	count, err = backfillWallets(ctx, tx, mongoDB.Collection("wallets"), opts.batchSize)
	if err != nil {
		return stats, fmt.Errorf("backfill wallets: %w", err)
	}
	stats.wallets = count
	log.Info("backfilled wallets", zap.Int("count", count))

	if err := tx.Commit(); err != nil {
		return stats, err
	}
	return stats, nil
}

func backfillCustomers(ctx context.Context, tx *sql.Tx, collection *mongo.Collection, batchSize int32, deleted bool) (int, error) {
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(batchSize))
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return count, err
		}
		id := stringField(doc, "_id", "id")
		if id == "" {
			return count, fmt.Errorf("customer document without id")
		}
		updatedAt := timeField(doc, "updatedAt", "createdAt")
		if updatedAt.IsZero() {
			updatedAt = time.Now().UTC()
		}
		if deleted {
			deletedAt := timeField(doc, "deletedAt")
			if deletedAt.IsZero() {
				deletedAt = updatedAt
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO deleted_customers (id,name,surname,email,phone,note,deleted_at) VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (id) DO UPDATE SET name=$2,surname=$3,email=$4,phone=$5,note=$6,deleted_at=$7`, id, stringField(doc, "name"), stringField(doc, "surname"), nullableString(doc, "email"), nullableString(doc, "phone"), stringField(doc, "note"), deletedAt)
		} else {
			_, err = tx.ExecContext(ctx, `INSERT INTO customers (id,name,surname,email,phone,note,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT (id) DO UPDATE SET name=$2,surname=$3,email=$4,phone=$5,note=$6,updated_at=$8`, id, stringField(doc, "name"), stringField(doc, "surname"), nullableString(doc, "email"), nullableString(doc, "phone"), stringField(doc, "note"), updatedAt, updatedAt)
		}
		if err != nil {
			return count, err
		}
		count++
	}
	return count, cursor.Err()
}

func backfillFidelityCards(ctx context.Context, tx *sql.Tx, collection *mongo.Collection, batchSize int32) (int, error) {
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(batchSize))
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return count, err
		}
		id := stringField(doc, "_id", "id")
		if id == "" {
			return count, fmt.Errorf("fidelity card document without id")
		}
		vouchers, err := normalizeVouchers(doc["vouchers"])
		if err != nil {
			return count, err
		}
		createdAt := timeField(doc, "createdAt", "updatedAt")
		updatedAt := timeField(doc, "updatedAt", "createdAt")
		_, err = tx.ExecContext(ctx, `INSERT INTO fidelity_cards (id,customer_id,solarium_purchases,vouchers,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (id) DO UPDATE SET customer_id=$2,solarium_purchases=$3,vouchers=$4,updated_at=$6`, id, stringField(doc, "customerId"), intField(doc, "solariumPurchases"), vouchers, zeroTimeToNow(createdAt), zeroTimeToNow(updatedAt))
		if err != nil {
			return count, err
		}
		count++
	}
	return count, cursor.Err()
}

func backfillWallets(ctx context.Context, tx *sql.Tx, collection *mongo.Collection, batchSize int32) (int, error) {
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(batchSize))
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return count, err
		}
		id := stringField(doc, "_id", "id")
		if id == "" {
			return count, fmt.Errorf("wallet document without id")
		}
		owner := stringField(doc, "owner")
		operations, err := normalizeWalletOperations(doc["operations"])
		if err != nil {
			return count, err
		}
		giftCards, err := normalizeGiftCards(doc["activeGiftCards"], owner)
		if err != nil {
			return count, err
		}
		createdAt := timeField(doc, "createdAt", "updatedAt")
		updatedAt := timeField(doc, "updatedAt", "createdAt")
		_, err = tx.ExecContext(ctx, `INSERT INTO wallets (id,owner,available_amount,spent,operations,gift_cards,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT (id) DO UPDATE SET owner=$2,available_amount=$3,spent=$4,operations=$5,gift_cards=$6,updated_at=$8`, id, owner, floatField(doc, "availableAmount"), floatField(doc, "spentAmount"), operations, giftCards, zeroTimeToNow(createdAt), zeroTimeToNow(updatedAt))
		if err != nil {
			return count, err
		}
		count++
	}
	return count, cursor.Err()
}

func normalizeVouchers(value any) (string, error) {
	items := documents(value)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"id":        stringField(item, "_id", "id"),
			"issuedAt":  timeField(item, "issuedAt", "createdAt"),
			"isUsed":    boolField(item, "isUsed"),
			"treatment": stringField(item, "treatment"),
		})
	}
	return marshalJSON(out)
}

func normalizeWalletOperations(value any) (string, error) {
	items := documents(value)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		typeName := stringField(item, "type", "_type")
		operation := map[string]any{"at": timeField(item, "at"), "amount": floatField(item, "amount")}
		switch strings.ToLower(typeName) {
		case "moneycredited", "money_credited", "moneycreditedentity":
			operation["type"] = "moneyCredited"
		case "giftcardmoneycredited", "gift_card_money_credited", "giftcardmoneycreditedentity":
			operation["type"] = "giftCardMoneyCredited"
			operation["giftCardId"] = stringField(item, "giftCardId")
			operation["expireAt"] = timeField(item, "expireAt", "expiresAt")
		case "giftcardmoneyexpired", "gift_card_money_expired", "giftcardmoneyexpiredentity":
			operation["type"] = "giftCardMoneyExpired"
			operation["giftCardId"] = stringField(item, "giftCardId")
		case "moneycharged", "moneycharge", "money_charged", "moneychargedentity":
			operation["type"] = "moneyCharged"
		default:
			operation["type"] = typeName
		}
		out = append(out, operation)
	}
	return marshalJSON(out)
}

func normalizeGiftCards(value any, owner string) (string, error) {
	items := documents(value)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		cardOwner := stringField(item, "customerId", "owner")
		if cardOwner == "" {
			cardOwner = owner
		}
		out = append(out, map[string]any{
			"id":              stringField(item, "_id", "id"),
			"owner":           cardOwner,
			"availableAmount": floatField(item, "availableAmount"),
			"createdAt":       timeField(item, "createdAt"),
			"expiresAt":       timeField(item, "expiresAt", "expireAt"),
			"amountSpent":     floatField(item, "amountSpent"),
		})
	}
	return marshalJSON(out)
}

func documents(value any) []bson.M {
	switch typed := value.(type) {
	case primitive.A:
		out := make([]bson.M, 0, len(typed))
		for _, item := range typed {
			if doc, ok := asDocument(item); ok {
				out = append(out, doc)
			}
		}
		return out
	case []any:
		out := make([]bson.M, 0, len(typed))
		for _, item := range typed {
			if doc, ok := asDocument(item); ok {
				out = append(out, doc)
			}
		}
		return out
	default:
		return nil
	}
}

func asDocument(value any) (bson.M, bool) {
	switch typed := value.(type) {
	case bson.M:
		return typed, true
	case bson.D:
		return typed.Map(), true
	default:
		return nil, false
	}
}

func marshalJSON(value any) (string, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func stringField(doc bson.M, keys ...string) string {
	for _, key := range keys {
		if value, ok := doc[key]; ok && value != nil {
			switch typed := value.(type) {
			case string:
				return typed
			case primitive.ObjectID:
				return typed.Hex()
			default:
				return fmt.Sprint(typed)
			}
		}
	}
	return ""
}

func nullableString(doc bson.M, key string) any {
	value := strings.TrimSpace(stringField(doc, key))
	if value == "" {
		return nil
	}
	return value
}

func intField(doc bson.M, key string) int {
	switch value := doc[key].(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	case float64:
		return int(value)
	case float32:
		return int(value)
	default:
		return 0
	}
}

func floatField(doc bson.M, key string) float64 {
	switch value := doc[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	default:
		return 0
	}
}

func boolField(doc bson.M, key string) bool {
	value, _ := doc[key].(bool)
	return value
}

func timeField(doc bson.M, keys ...string) time.Time {
	for _, key := range keys {
		if value, ok := doc[key]; ok && value != nil {
			switch typed := value.(type) {
			case time.Time:
				return typed.UTC()
			case primitive.DateTime:
				return typed.Time().UTC()
			}
		}
	}
	return time.Time{}
}

func zeroTimeToNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value
}
