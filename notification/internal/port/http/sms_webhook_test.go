package http

import (
	"context"
	"testing"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api/smswebhook"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/domain"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/infra/provider"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestSmsGatewayNotifyNotificationNotFoundLogsWarning(t *testing.T) {
	core, observedLogs := observer.New(zap.WarnLevel)
	service := application.NewNotificationService(notificationRepositoryStub{}, notificationProviderStub{})
	server := NewSmsHandler(service, zap.New(core))

	eventType := smswebhook.MessageDeliverSucceeded
	metadata := smswebhook.AdditionalMetadata{provider.NotificationIDMetadata: "missing-notification"}

	response, err := server.SmsGatewayNotify(context.Background(), smswebhook.SmsGatewayNotifyRequestObject{
		Body: &smswebhook.SmsGatewayNotifyJSONRequestBody{
			EventType: &eventType,
			Metadata:  &metadata,
		},
	})
	if err != nil {
		t.Fatalf("SmsGatewayNotify returned error: %v", err)
	}
	if _, ok := response.(smswebhook.SmsGatewayNotify200Response); !ok {
		t.Fatalf("response = %T, want SmsGatewayNotify200Response", response)
	}

	warnings := observedLogs.FilterMessage("notification not found while confirming sms delivery").All()
	if len(warnings) != 1 {
		t.Fatalf("warning logs = %d, want 1", len(warnings))
	}
	if observedLogs.FilterMessage("failed to confirm notification sent").Len() != 0 {
		t.Fatal("unexpected error log for missing notification")
	}
}

type notificationRepositoryStub struct{}

func (notificationRepositoryStub) FindByID(context.Context, string) (*domain.Notification, error) {
	return nil, nil
}

func (notificationRepositoryStub) Save(context.Context, *domain.Notification) error {
	return nil
}

type notificationProviderStub struct{}

func (notificationProviderStub) Supports(domain.Notification) bool {
	return true
}

func (notificationProviderStub) Send(context.Context, domain.Notification) (domain.ChannelMetadata, error) {
	return domain.ChannelMetadata{}, nil
}
