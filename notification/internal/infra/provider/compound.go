package provider

import (
	"context"
	"fmt"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/domain"
)

type NotificationProvider interface {
	Supports(notification domain.Notification) bool
	Send(ctx context.Context, notification domain.Notification) (domain.ChannelMetadata, error)
}

type CompoundProvider struct {
	providers []NotificationProvider
}

func NewCompoundProvider(providers ...NotificationProvider) *CompoundProvider {
	return &CompoundProvider{providers: providers}
}

func (provider *CompoundProvider) Supports(notification domain.Notification) bool {
	for _, candidate := range provider.providers {
		if candidate.Supports(notification) {
			return true
		}
	}
	return false
}

func (provider *CompoundProvider) Send(ctx context.Context, notification domain.Notification) (domain.ChannelMetadata, error) {
	for _, candidate := range provider.providers {
		if candidate.Supports(notification) {
			return candidate.Send(ctx, notification)
		}
	}
	return domain.ChannelMetadata{}, fmt.Errorf("no provider supports %s", notification.Channel.Type)
}
