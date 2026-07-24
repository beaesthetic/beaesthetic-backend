package http

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/api/smswebhook"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestSmsGatewayNotifyConfirmsCustomerNotificationDelivery(t *testing.T) {
	customerRepository := &customerNotificationRepositoryStub{sentMatch: true}
	customerService := application.NewCustomerNotificationService(nil, nil, customerRepository, nil)
	server := NewSmsWebhookHandler(customerService, zap.NewNop())

	eventType := smswebhook.MessageDeliverSucceeded
	smsGatewayMessageID := uuid.MustParse("044207cc-35bd-4027-9034-e07ec51b4635")

	response, err := server.SmsGatewayNotify(context.Background(), smswebhook.SmsGatewayNotifyRequestObject{
		Body: &smswebhook.SmsGatewayNotifyJSONRequestBody{
			EventType: &eventType,
			Data: &smswebhook.SmsEntityResponse{
				Id: smsGatewayMessageID,
			},
		},
	})
	if err != nil {
		t.Fatalf("SmsGatewayNotify returned error: %v", err)
	}
	if _, ok := response.(smswebhook.SmsGatewayNotify200Response); !ok {
		t.Fatalf("response = %T, want SmsGatewayNotify200Response", response)
	}
	if customerRepository.sentMessageID != smsGatewayMessageID.String() {
		t.Fatalf("sent sms gateway message id = %q, want %q", customerRepository.sentMessageID, smsGatewayMessageID.String())
	}
}

func TestSmsGatewayNotifyUnmatchedCustomerNotificationLogsWarning(t *testing.T) {
	core, observedLogs := observer.New(zap.WarnLevel)
	customerRepository := &customerNotificationRepositoryStub{sentMatch: false}
	customerService := application.NewCustomerNotificationService(nil, nil, customerRepository, nil)
	server := NewSmsWebhookHandler(customerService, zap.New(core))

	eventType := smswebhook.MessageDeliverSucceeded

	response, err := server.SmsGatewayNotify(context.Background(), smswebhook.SmsGatewayNotifyRequestObject{
		Body: &smswebhook.SmsGatewayNotifyJSONRequestBody{
			EventType: &eventType,
			Data: &smswebhook.SmsEntityResponse{
				Id: uuid.MustParse("044207cc-35bd-4027-9034-e07ec51b4635"),
			},
		},
	})
	if err != nil {
		t.Fatalf("SmsGatewayNotify returned error: %v", err)
	}
	if _, ok := response.(smswebhook.SmsGatewayNotify200Response); !ok {
		t.Fatalf("response = %T, want SmsGatewayNotify200Response", response)
	}
	warnings := observedLogs.FilterMessage("customer notification not found while confirming sms delivery").All()
	if len(warnings) != 1 {
		t.Fatalf("warning logs = %d, want 1", len(warnings))
	}
}

type customerNotificationRepositoryStub struct {
	sentMatch     bool
	failedMatch   bool
	sentMessageID string
}

func (repo *customerNotificationRepositoryStub) Exists(context.Context, string) (bool, error) {
	return false, nil
}

func (repo *customerNotificationRepositoryStub) CreatePending(context.Context, application.CustomerNotificationRecord) (bool, error) {
	return true, nil
}

func (repo *customerNotificationRepositoryStub) SaveSMSGatewayDispatch(context.Context, application.SMSGatewayDispatch) error {
	return nil
}

func (repo *customerNotificationRepositoryStub) MarkDispatched(context.Context, string, time.Time) error {
	return nil
}

func (repo *customerNotificationRepositoryStub) MarkFailed(context.Context, string, string, string, time.Time) (bool, error) {
	return repo.failedMatch, nil
}

func (repo *customerNotificationRepositoryStub) MarkSentBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, sentAt time.Time) (bool, error) {
	repo.sentMessageID = smsGatewayMessageID
	return repo.sentMatch, nil
}

func (repo *customerNotificationRepositoryStub) MarkFailedBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, reason string, message string, failedAt time.Time) (bool, error) {
	return repo.failedMatch, nil
}
