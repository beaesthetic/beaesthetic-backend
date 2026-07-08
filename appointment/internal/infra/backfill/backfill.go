package backfill

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func Run(ctx context.Context, cfg config.Config, log *zap.Logger) error {
	db, err := pgxpool.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}
	defer client.Disconnect(ctx)
	mongoDatabase, err := mongoDatabaseName(cfg.Mongo.URI, cfg.Mongo.Database)
	if err != nil {
		return err
	}
	log.Info("starting appointment backfill", zap.String("mongo_database", mongoDatabase))
	mongoDB := client.Database(mongoDatabase)
	agenda, err := backfillAgenda(ctx, db, mongoDB.Collection("agenda"))
	if err != nil {
		return err
	}
	services, err := backfillServices(ctx, db, mongoDB.Collection("services"))
	if err != nil {
		return err
	}
	log.Info("backfill completed", zap.Int("agenda", agenda), zap.Int("services", services))
	return nil
}

func mongoDatabaseName(uri string, configured string) (string, error) {
	configured = strings.TrimSpace(configured)
	if configured != "" {
		return configured, nil
	}

	parsed, err := url.Parse(uri)
	if err == nil {
		if db := strings.Trim(strings.TrimSpace(parsed.Path), "/"); db != "" {
			return db, nil
		}
	}

	return "", fmt.Errorf("mongo database is required: set ENV_MONGO_DATABASE or include a database path in ENV_MONGO_URI")
}

type agendaDoc struct {
	ID       string    `bson:"_id"`
	Start    time.Time `bson:"start"`
	End      time.Time `bson:"end"`
	Attendee struct {
		ID          string `bson:"id"`
		DisplayName string `bson:"displayName"`
	} `bson:"attendee"`
	Data                bson.Raw   `bson:"data"`
	CancelReason        *string    `bson:"cancelReason"`
	RemindBeforeSeconds int        `bson:"remindBeforeSeconds"`
	ReminderStatus      string     `bson:"reminderStatus"`
	ReminderSentAt      *time.Time `bson:"reminderSentAt"`
	Version             int64      `bson:"version"`
	CreatedAt           time.Time  `bson:"createdAt"`
	UpdatedAt           time.Time  `bson:"updatedAt"`
}

func backfillAgenda(ctx context.Context, db *pgxpool.Pool, col *mongo.Collection) (int, error) {
	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx)
	count := 0
	for cur.Next(ctx) {
		var d agendaDoc
		if err := cur.Decode(&d); err != nil {
			return count, err
		}
		typ, title, description, services, err := parseAgendaData(d.Data)
		if err != nil {
			return count, err
		}
		servicesJSON, _ := json.Marshal(services)
		_, err = db.Exec(ctx, `INSERT INTO agenda_events (id,event_type,title,description,start_at,end_at,attendee_id,attendee_display_name,services,cancel_reason,reminder_status,reminder_sent_at,remind_before_seconds,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16) ON CONFLICT (id) DO NOTHING`, d.ID, typ, title, description, d.Start, d.End, d.Attendee.ID, d.Attendee.DisplayName, servicesJSON, d.CancelReason, d.ReminderStatus, d.ReminderSentAt, d.RemindBeforeSeconds, d.Version, d.CreatedAt, d.UpdatedAt)
		if err != nil {
			return count, fmt.Errorf("insert agenda %s: %w", d.ID, err)
		}
		count++
	}
	return count, cur.Err()
}

func parseAgendaData(raw bson.Raw) (string, string, string, []map[string]string, error) {
	var probe struct {
		Type        string   `bson:"type"`
		Services    []string `bson:"services"`
		Title       string   `bson:"title"`
		Description string   `bson:"description"`
	}
	if err := bson.Unmarshal(raw, &probe); err != nil {
		return "", "", "", nil, err
	}
	if len(probe.Services) > 0 {
		services := make([]map[string]string, 0, len(probe.Services))
		for _, name := range probe.Services {
			services = append(services, map[string]string{"Name": name})
		}
		return "appointment", probe.Title, probe.Description, services, nil
	}
	return "event", probe.Title, probe.Description, nil, nil
}

type serviceDoc struct {
	ID          string   `bson:"_id"`
	Name        string   `bson:"name"`
	Price       float64  `bson:"price"`
	Tags        []string `bson:"tags"`
	ColorHex    *string  `bson:"colorHex"`
	SearchGrams string   `bson:"searchGrams"`
}

func backfillServices(ctx context.Context, db *pgxpool.Pool, col *mongo.Collection) (int, error) {
	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	defer cur.Close(ctx)
	count := 0
	for cur.Next(ctx) {
		var d serviceDoc
		if err := cur.Decode(&d); err != nil {
			return count, err
		}
		tags, _ := json.Marshal(d.Tags)
		_, err = db.Exec(ctx, `INSERT INTO appointment_services (id,name,price,tags,color_hex) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (id) DO NOTHING`, d.ID, d.Name, d.Price, tags, d.ColorHex)
		if err != nil {
			return count, fmt.Errorf("insert service %s: %w", d.ID, err)
		}
		count++
	}
	return count, cur.Err()
}
