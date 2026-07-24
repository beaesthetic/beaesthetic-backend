package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCustomerNotificationServiceProcessSendsSms(t *testing.T) {
	provider := &fakeSMSDispatcher{}
	templates := &fakeTemplateRenderer{content: "hello Ada"}
	customerNotifications := newFakeCustomerNotificationRepository()
	customerNotifications.dispatcher = provider
	service := NewCustomerNotificationService(
		fakeCustomerReader{"customer-1": {ID: "customer-1", Name: "Ada", Surname: "Lovelace", Phone: "+393331234567"}},
		templates,
		customerNotifications,
		provider,
	)
	service.now = func() time.Time { return time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC) }

	err := service.Process(context.Background(), CustomerNotificationCommand{
		IdempotencyKey:      "external-key",
		CustomerIDs:         []string{"customer-1"},
		NotificationChannel: "sms",
		NotificationType:    "appointment_reminder",
		Body:                map[string]any{"date": "2026-07-20"},
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if provider.sent != 1 {
		t.Fatalf("sent notifications = %d, want 1", provider.sent)
	}
	if !customerNotifications.createdBeforeSend {
		t.Fatal("customer notification should be created before sending to provider")
	}
	if customerNotifications.created == nil {
		t.Fatal("customer notification should be stored")
	}
	if customerNotifications.created.TemplateValues["date"] != "2026-07-20" {
		t.Fatalf("stored template date = %v, want 2026-07-20", customerNotifications.created.TemplateValues["date"])
	}
	if _, ok := customerNotifications.created.TemplateValues["name"]; ok {
		t.Fatal("stored template values should come from queue body, not customer enrichment")
	}
	if templates.data.Values["name"] != "Ada" {
		t.Fatalf("template name = %v, want Ada", templates.data.Values["name"])
	}
	if templates.data.Values["date"] != "2026-07-20" {
		t.Fatalf("template date = %v, want 2026-07-20", templates.data.Values["date"])
	}
}

func TestCustomerNotificationServiceSkipsExistingIdempotencyKey(t *testing.T) {
	provider := &fakeSMSDispatcher{}
	customerNotifications := newFakeCustomerNotificationRepository()
	customerNotifications.keys["external-key:customer-1:sms:appointment_reminder"] = true
	service := NewCustomerNotificationService(
		fakeCustomerReader{"customer-1": {ID: "customer-1", Phone: "+393331234567"}},
		&fakeTemplateRenderer{content: "hello"},
		customerNotifications,
		provider,
	)

	err := service.Process(context.Background(), CustomerNotificationCommand{
		IdempotencyKey:      "external-key",
		CustomerIDs:         []string{"customer-1"},
		NotificationChannel: "sms",
		NotificationType:    "appointment_reminder",
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if provider.sent != 0 {
		t.Fatalf("sent notifications = %d, want 0", provider.sent)
	}
}

func TestCustomerNotificationServiceRequiresPhone(t *testing.T) {
	repo := newFakeCustomerNotificationRepository()
	service := NewCustomerNotificationService(
		fakeCustomerReader{"customer-1": {ID: "customer-1"}},
		&fakeTemplateRenderer{content: "hello"},
		repo,
		&fakeSMSDispatcher{},
	)

	err := service.Process(context.Background(), CustomerNotificationCommand{
		IdempotencyKey:      "external-key",
		CustomerIDs:         []string{"customer-1"},
		NotificationChannel: "sms",
		NotificationType:    "appointment_reminder",
	})
	if err != nil {
		t.Fatalf("Process() error = %v, want nil", err)
	}
	if repo.failed == nil {
		t.Fatal("customer notification should be marked failed")
	}
	if repo.failed.reason != CustomerNotificationReasonCustomerPhoneRequired {
		t.Fatalf("failed reason = %q, want %q", repo.failed.reason, CustomerNotificationReasonCustomerPhoneRequired)
	}
}

func TestCustomerNotificationCommandTemplateValuesMergesCustomerAndBody(t *testing.T) {
	values := CustomerNotificationCommand{
		NotificationChannel: "sms",
		NotificationType:    "appointment_reminder",
		Body: map[string]any{
			"date": "2026-07-20",
			"name": "Body Name",
		},
	}.TemplateValues(Customer{ID: "customer-1", Name: "Ada", Surname: "Lovelace", Email: "ada@example.com", Phone: "+393331234567", Note: "vip"})

	if values["id"] != "customer-1" || values["surname"] != "Lovelace" || values["phone"] != "+393331234567" {
		t.Fatalf("customer values not merged correctly: %+v", values)
	}
	if values["date"] != "2026-07-20" {
		t.Fatalf("body date = %v, want 2026-07-20", values["date"])
	}
	if values["name"] != "Body Name" {
		t.Fatalf("body should override duplicated customer key, got %v", values["name"])
	}
}

func TestCustomerNotificationCommandRejectsEmailForNow(t *testing.T) {
	err := CustomerNotificationCommand{
		CustomerIDs:         []string{"customer-1"},
		NotificationChannel: "email",
		NotificationType:    "appointment_reminder",
	}.Validate()
	if !errors.Is(err, ErrUnsupportedCustomerNotificationChannel) {
		t.Fatalf("Validate() error = %v, want ErrUnsupportedCustomerNotificationChannel", err)
	}
}

type fakeCustomerReader map[string]Customer

func (reader fakeCustomerReader) GetCustomer(ctx context.Context, id string) (Customer, error) {
	customer, ok := reader[id]
	if !ok {
		return Customer{}, errors.New("customer not found")
	}
	return customer, nil
}

type fakeTemplateRenderer struct {
	content string
	data    CustomerNotificationTemplateData
}

func (renderer *fakeTemplateRenderer) Render(ctx context.Context, data CustomerNotificationTemplateData) (string, error) {
	renderer.data = data
	return renderer.content, nil
}

type fakeCustomerNotificationRepository struct {
	keys              map[string]bool
	createdBeforeSend bool
	dispatcher        *fakeSMSDispatcher
	created           *CustomerNotificationRecord
	dispatched        string
	failed            *fakeFailure
}

func newFakeCustomerNotificationRepository() *fakeCustomerNotificationRepository {
	return &fakeCustomerNotificationRepository{keys: map[string]bool{}}
}

func (repo *fakeCustomerNotificationRepository) Exists(ctx context.Context, idempotencyKey string) (bool, error) {
	return repo.keys[idempotencyKey], nil
}

func (repo *fakeCustomerNotificationRepository) CreatePending(ctx context.Context, delivery CustomerNotificationRecord) (bool, error) {
	repo.keys[delivery.IdempotencyKey] = true
	copy := delivery
	repo.created = &copy
	if repo.dispatcher != nil && repo.dispatcher.sent == 0 {
		repo.createdBeforeSend = true
	}
	return true, nil
}

func (repo *fakeCustomerNotificationRepository) SaveSMSGatewayDispatch(ctx context.Context, dispatch SMSGatewayDispatch) error {
	return nil
}

func (repo *fakeCustomerNotificationRepository) MarkDispatched(ctx context.Context, notificationID string, dispatchedAt time.Time) error {
	repo.dispatched = notificationID
	return nil
}

func (repo *fakeCustomerNotificationRepository) MarkFailed(ctx context.Context, notificationID string, reason string, message string, failedAt time.Time) (bool, error) {
	repo.failed = &fakeFailure{notificationID: notificationID, reason: reason, message: message}
	return true, nil
}

func (repo *fakeCustomerNotificationRepository) MarkSentBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, sentAt time.Time) (bool, error) {
	return true, nil
}

func (repo *fakeCustomerNotificationRepository) MarkFailedBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, reason string, message string, failedAt time.Time) (bool, error) {
	return true, nil
}

type fakeFailure struct {
	notificationID string
	reason         string
	message        string
}

type fakeSMSDispatcher struct{ sent int }

func (dispatcher *fakeSMSDispatcher) Send(ctx context.Context, messageID, phone, content string) (string, error) {
	dispatcher.sent++
	return "provider-id", nil
}
