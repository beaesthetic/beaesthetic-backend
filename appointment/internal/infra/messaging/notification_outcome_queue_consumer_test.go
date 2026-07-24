package messaging

import (
	"testing"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	notificationcontracts "github.com/petretiandrea/beaesthetic-backend/core-contracts/notification"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestNotificationOutcomeHandlerMapsFailedStatus(t *testing.T) {
	consumer := &NotificationOutcomeQueueConsumer{service: &application.AppointmentService{}}
	handle, status, err := consumer.notificationOutcomeHandler(notificationcontracts.CustomerNotificationOutcomeStatus_CUSTOMER_NOTIFICATION_OUTCOME_STATUS_FAILED)
	if err != nil {
		t.Fatalf("notificationOutcomeHandler() error = %v", err)
	}
	if handle == nil {
		t.Fatal("expected handler")
	}
	if status != "failed" {
		t.Fatalf("status = %q, want failed", status)
	}
}

func TestNotificationOutcomePayloadUsesCoreContract(t *testing.T) {
	payload, err := protojson.Marshal(&notificationcontracts.CustomerNotificationOutcome{
		NotificationId: "notification-1",
		Status:         notificationcontracts.CustomerNotificationOutcomeStatus_CUSTOMER_NOTIFICATION_OUTCOME_STATUS_FAILED,
		Reason:         "missing_customer_contact",
		Message:        "customer phone is required",
		IdempotencyKey: "notification-1:customer-1:sms:appointment_reminder",
		CustomerId:     "customer-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	var event notificationcontracts.CustomerNotificationOutcome
	if err := protojson.Unmarshal(payload, &event); err != nil {
		t.Fatalf("protojson.Unmarshal() error = %v", err)
	}
	if event.GetNotificationId() != "notification-1" || event.GetReason() != "missing_customer_contact" || event.GetIdempotencyKey() == "" || event.GetCustomerId() != "customer-1" {
		t.Fatalf("unexpected outcome: %+v", event)
	}
}
