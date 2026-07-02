package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/domain"
)

type NotificationRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Notification, error)
	Save(ctx context.Context, notification *domain.Notification) error
}

type NotificationProvider interface {
	Supports(notification domain.Notification) bool
	Send(ctx context.Context, notification domain.Notification) (domain.ChannelMetadata, error)
}

type NotificationService struct {
	repository NotificationRepository
	provider   NotificationProvider
	now        func() time.Time
}

func NewNotificationService(repository NotificationRepository, provider NotificationProvider) *NotificationService {
	return &NotificationService{repository: repository, provider: provider, now: time.Now}
}

func (service *NotificationService) CreateNotification(ctx context.Context, title, content string, channel domain.Channel) (*domain.Notification, error) {
	notification, err := domain.NewNotification(uuid.NewString(), title, content, channel, service.now())
	if err != nil {
		return nil, err
	}
	if err := service.repository.Save(ctx, &notification); err != nil {
		return nil, err
	}
	return &notification, nil
}

func (service *NotificationService) GetNotification(ctx context.Context, id string) (*domain.Notification, error) {
	return service.repository.FindByID(ctx, id)
}

func (service *NotificationService) SendNotification(ctx context.Context, id string) error {
	notification, err := service.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if notification == nil {
		return fmt.Errorf("notification %s not found", id)
	}
	if !service.provider.Supports(*notification) {
		return fmt.Errorf("provider for %s is not supported", notification.Channel.Type)
	}
	metadata, err := service.provider.Send(ctx, *notification)
	if err != nil {
		return err
	}
	notification.MarkSent(metadata, service.now())
	return service.repository.Save(ctx, notification)
}

func (service *NotificationService) ConfirmNotificationSent(ctx context.Context, id string) error {
	notification, err := service.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if notification == nil {
		return fmt.Errorf("notification %s not found", id)
	}
	notification.ConfirmSent(service.now())
	return service.repository.Save(ctx, notification)
}
