package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

type SchedulerClient struct {
	client *ClientWithResponses
}

func NewSchedulerClient(baseURL string) (*SchedulerClient, error) {
	client, err := NewClientWithResponses(baseURL)
	if err != nil {
		return nil, err
	}
	return &SchedulerClient{client: client}, nil
}

func (c *SchedulerClient) Raw() *ClientWithResponses {
	return c.client
}

func (c *SchedulerClient) Schedule(ctx context.Context, scheduleID string, scheduleAt time.Time, route string, data map[string]any) error {
	payload := map[string]any{
		"scheduleAt": scheduleAt.UTC(),
		"route":      route,
		"data":       data,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal schedule payload: %w", err)
	}
	parsedID, err := uuid.Parse(scheduleID)
	if err != nil {
		return fmt.Errorf("parse schedule id: %w", err)
	}
	response, err := c.client.AddScheduleWithBodyWithResponse(ctx, openapi_types.UUID(parsedID), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	if response.StatusCode() != 202 {
		return fmt.Errorf("schedule request rejected with status %s", response.Status())
	}
	return nil
}

func (c *SchedulerClient) Unschedule(ctx context.Context, scheduleID string) error {
	parsedID, err := uuid.Parse(scheduleID)
	if err != nil {
		return fmt.Errorf("parse schedule id: %w", err)
	}
	response, err := c.client.RemoveScheduleWithResponse(ctx, openapi_types.UUID(parsedID))
	if err != nil {
		return err
	}
	if response.StatusCode() != 204 {
		return fmt.Errorf("unschedule request rejected with status %s", response.Status())
	}
	return nil
}

type ReminderScheduler struct {
	client *SchedulerClient
	route  string
}

func NewReminderScheduler(client *SchedulerClient, route string) *ReminderScheduler {
	return &ReminderScheduler{client: client, route: route}
}

func (s *ReminderScheduler) ScheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent, sendAt time.Time) error {
	return s.client.Schedule(ctx, agendaEvent.ID, sendAt, s.route, map[string]any{"eventId": agendaEvent.ID})
}

func (s *ReminderScheduler) UnscheduleReminder(ctx context.Context, eventID string) error {
	return s.client.Unschedule(ctx, eventID)
}
