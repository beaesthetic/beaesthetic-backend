package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/domain"
)

func TestCustomerNotificationServiceProcessSendsSms(t *testing.T) {
	repo := newFakeNotificationRepository()
	provider := &fakeNotificationProvider{}
	notifications := NewNotificationService(repo, provider)
	templates := &fakeTemplateRenderer{content: "hello Ada"}
	service := NewCustomerNotificationService(
		fakeCustomerReader{"customer-1": {ID: "customer-1", Name: "Ada", Surname: "Lovelace", Phone: "+393331234567"}},
		templates,
		newFakeIdempotencyRepository(),
		notifications,
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
	if len(repo.notifications) != 1 {
		t.Fatalf("saved notifications = %d, want 1", len(repo.notifications))
	}
	if templates.data.Values["name"] != "Ada" {
		t.Fatalf("template name = %v, want Ada", templates.data.Values["name"])
	}
	if templates.data.Values["date"] != "2026-07-20" {
		t.Fatalf("template date = %v, want 2026-07-20", templates.data.Values["date"])
	}
}

func TestCustomerNotificationServiceSkipsExistingIdempotencyKey(t *testing.T) {
	repo := newFakeNotificationRepository()
	provider := &fakeNotificationProvider{}
	idempotency := newFakeIdempotencyRepository()
	idempotency.keys["external-key:customer-1:sms:appointment_reminder"] = true
	service := NewCustomerNotificationService(
		fakeCustomerReader{"customer-1": {ID: "customer-1", Phone: "+393331234567"}},
		&fakeTemplateRenderer{content: "hello"},
		idempotency,
		NewNotificationService(repo, provider),
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
	service := NewCustomerNotificationService(
		fakeCustomerReader{"customer-1": {ID: "customer-1"}},
		&fakeTemplateRenderer{content: "hello"},
		newFakeIdempotencyRepository(),
		NewNotificationService(newFakeNotificationRepository(), &fakeNotificationProvider{}),
	)

	err := service.Process(context.Background(), CustomerNotificationCommand{
		IdempotencyKey:      "external-key",
		CustomerIDs:         []string{"customer-1"},
		NotificationChannel: "sms",
		NotificationType:    "appointment_reminder",
	})
	if !errors.Is(err, ErrCustomerPhoneRequired) {
		t.Fatalf("Process() error = %v, want ErrCustomerPhoneRequired", err)
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

type fakeIdempotencyRepository struct{ keys map[string]bool }

func newFakeIdempotencyRepository() *fakeIdempotencyRepository {
	return &fakeIdempotencyRepository{keys: map[string]bool{}}
}

func (repo *fakeIdempotencyRepository) Exists(ctx context.Context, idempotencyKey string) (bool, error) {
	return repo.keys[idempotencyKey], nil
}

func (repo *fakeIdempotencyRepository) Save(ctx context.Context, delivery CustomerNotificationDelivery) error {
	repo.keys[delivery.IdempotencyKey] = true
	return nil
}

type fakeNotificationRepository struct {
	notifications map[string]*domain.Notification
}

func newFakeNotificationRepository() *fakeNotificationRepository {
	return &fakeNotificationRepository{notifications: map[string]*domain.Notification{}}
}

func (repo *fakeNotificationRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	return repo.notifications[id], nil
}

func (repo *fakeNotificationRepository) Save(ctx context.Context, notification *domain.Notification) error {
	copy := *notification
	repo.notifications[notification.ID] = &copy
	return nil
}

type fakeNotificationProvider struct{ sent int }

func (provider *fakeNotificationProvider) Supports(notification domain.Notification) bool {
	return notification.Channel.Type == domain.ChannelSMS
}

func (provider *fakeNotificationProvider) Send(ctx context.Context, notification domain.Notification) (domain.ChannelMetadata, error) {
	provider.sent++
	return domain.ChannelMetadata{ProviderResourceID: "provider-id"}, nil
}
