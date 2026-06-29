package redislegacy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beaesthetic/scheduler/internal/config"
	"github.com/beaesthetic/scheduler/internal/domain"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Store interface {
	Save(ctx context.Context, job domain.ScheduleJob) error
}

type Migrator struct {
	config *config.Config
	store  Store
	logger *zap.Logger
}

type Result struct {
	Copied  int
	Skipped int
	Failed  int
}

type legacyJob struct {
	ID   legacyID `json:"id"`
	Meta struct {
		Route string         `json:"route"`
		Data  map[string]any `json:"data"`
	} `json:"meta"`
	ScheduleAt time.Time `json:"scheduleAt"`
}

type legacyID string

func (id *legacyID) UnmarshalJSON(data []byte) error {
	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		*id = legacyID(asString)
		return nil
	}

	var asObject struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &asObject); err != nil {
		return err
	}
	*id = legacyID(asObject.ID)
	return nil
}

func NewMigrator(cfg *config.Config, store Store, logger *zap.Logger) *Migrator {
	return &Migrator{config: cfg, store: store, logger: logger}
}

func (m *Migrator) Run(ctx context.Context, dryRun bool) (Result, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", m.config.Redis.Host, m.config.Redis.Port),
		Password: m.config.Redis.Password,
		DB:       m.config.Redis.DB,
	})
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		return Result{}, err
	}

	key := fmt.Sprintf("%s-clock", m.config.Scheduler.Name)
	ids, err := client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return Result{}, err
	}

	result := Result{}
	for _, rawID := range ids {
		raw, err := client.Get(ctx, rawID).Result()
		if err != nil {
			result.Failed++
			m.logger.Error("failed to read legacy schedule", zap.String("schedule_id", rawID), zap.Error(err))
			continue
		}

		job, err := parseLegacyJob(raw)
		if err != nil {
			result.Failed++
			m.logger.Error("failed to parse legacy schedule", zap.String("schedule_id", rawID), zap.Error(err))
			continue
		}

		if dryRun {
			result.Skipped++
			continue
		}

		if err := m.store.Save(ctx, job); err != nil {
			result.Failed++
			m.logger.Error("failed to copy legacy schedule", zap.String("schedule_id", rawID), zap.Error(err))
			continue
		}
		result.Copied++
	}

	return result, nil
}

func parseLegacyJob(raw string) (domain.ScheduleJob, error) {
	var legacy legacyJob
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return domain.ScheduleJob{}, err
	}

	id, err := uuid.Parse(string(legacy.ID))
	if err != nil {
		return domain.ScheduleJob{}, err
	}
	if legacy.Meta.Data == nil {
		legacy.Meta.Data = map[string]any{}
	}

	return domain.ScheduleJob{
		ID:         id,
		Route:      legacy.Meta.Route,
		Payload:    legacy.Meta.Data,
		ScheduleAt: legacy.ScheduleAt,
	}, nil
}
