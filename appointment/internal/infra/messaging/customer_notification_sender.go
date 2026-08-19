package messaging

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
	domainv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
	notification "github.com/petretiandrea/beaesthetic-backend/core-contracts/notification"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const ChannelCustomerNotifications = "customer.notifications"

type CustomerNotificationSender struct {
	publisher outbox.Publisher
}

func NewCustomerNotificationSender(publisher outbox.Publisher) *CustomerNotificationSender {
	return &CustomerNotificationSender{publisher: publisher}
}

func (sender *CustomerNotificationSender) SendAppointmentReminder(ctx context.Context, agendaEvent *domain.AgendaEvent) (string, error) {
	return sender.sendAppointmentNotification(ctx, agendaEvent, application.NotificationTypeAppointmentReminder)
}

func (sender *CustomerNotificationSender) SendCalendarNotification(ctx context.Context, event domainv2.CalendarEvent, notificationType domainv2.NotificationType, idempotencyKey string) (string, error) {
	appointment, ok := event.Detail.(domainv2.Appointment)
	if !ok {
		return "", fmt.Errorf("calendar event %s is not an appointment", event.ID)
	}
	body, err := structpb.NewStruct(map[string]any{
		"eventId": event.ID,
		"startAt": event.Range.Start.UTC().Format(time.RFC3339),
		"endAt":   event.Range.End.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return "", fmt.Errorf("build customer notification body: %w", err)
	}
	return sender.publishCustomerNotification(ctx, &notification.CustomerNotificationRequested{
		IdempotencyKey:      idempotencyKey,
		CustomerIds:         []string{appointment.Customer.ID},
		NotificationChannel: notification.NotificationChannel_NOTIFICATION_CHANNEL_SMS,
		NotificationType:    string(notificationType),
		Body:                body,
	})
}

func (sender *CustomerNotificationSender) SendAppointmentConfirmation(ctx context.Context, agendaEvent *domain.AgendaEvent) (string, error) {
	return sender.sendAppointmentNotification(ctx, agendaEvent, application.NotificationTypeAppointmentConfirmation)
}

func (sender *CustomerNotificationSender) SendAppointmentRescheduled(ctx context.Context, agendaEvent *domain.AgendaEvent) (string, error) {
	return sender.sendAppointmentNotification(ctx, agendaEvent, application.NotificationTypeAppointmentRescheduled)
}

func (sender *CustomerNotificationSender) sendAppointmentNotification(ctx context.Context, agendaEvent *domain.AgendaEvent, notificationType string) (string, error) {
	request, err := newCustomerNotificationRequest(agendaEvent, notificationType)
	if err != nil {
		return "", err
	}
	return sender.publishCustomerNotification(ctx, request)
}

func (sender *CustomerNotificationSender) publishCustomerNotification(ctx context.Context, request *notification.CustomerNotificationRequested) (string, error) {
	if request == nil {
		return "", fmt.Errorf("customer notification request is required")
	}
	if len(request.GetCustomerIds()) == 0 || strings.TrimSpace(request.GetCustomerIds()[0]) == "" {
		return "", fmt.Errorf("customerIds is required")
	}
	if strings.TrimSpace(request.GetNotificationType()) == "" {
		return "", fmt.Errorf("notificationType is required")
	}
	idempotencyKey := strings.TrimSpace(request.GetIdempotencyKey())
	if idempotencyKey == "" {
		idempotencyKey = uuid.NewString()
		request.IdempotencyKey = idempotencyKey
	}
	payload, err := protojson.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("marshal customer notification request: %w", err)
	}
	customerID := request.GetCustomerIds()[0]
	if err := sender.publisher.Publish(ctx, outbox.Message{
		ID:          uuid.NewString(),
		Channel:     outbox.Channel(ChannelCustomerNotifications),
		AffinityKey: outbox.AffinityKey(customerID),
		Payload:     payload,
		Metadata:    outbox.Metadata{},
		OccurredAt:  time.Now().UTC(),
	}); err != nil {
		return "", fmt.Errorf("publish customer notification request: %w", err)
	}
	return idempotencyKey, nil
}

func newCustomerNotificationRequest(event *domain.AgendaEvent, notificationType string) (*notification.CustomerNotificationRequested, error) {
	if event == nil {
		return nil, fmt.Errorf("agenda event is required")
	}
	body, err := structpb.NewStruct(map[string]any{
		"eventId": event.ID,
		"startAt": event.Start.UTC().Format(time.RFC3339),
		"endAt":   event.End.UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, fmt.Errorf("build customer notification body: %w", err)
	}
	return &notification.CustomerNotificationRequested{
		IdempotencyKey:      customerNotificationIdempotencyKey(event, notificationType),
		CustomerIds:         []string{event.Attendee.ID},
		NotificationChannel: notification.NotificationChannel_NOTIFICATION_CHANNEL_SMS,
		NotificationType:    notificationType,
		Body:                body,
	}, nil
}

func customerNotificationIdempotencyKey(event *domain.AgendaEvent, notificationType string) string {
	return fmt.Sprintf("appointment:%s:%s:%s", event.ID, notificationType, event.Start.UTC().Format(time.RFC3339))
}
