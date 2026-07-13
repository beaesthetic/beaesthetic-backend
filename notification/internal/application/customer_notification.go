package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/domain"
)

var (
	ErrUnsupportedCustomerNotificationChannel = errors.New("unsupported customer notification channel")
	ErrCustomerPhoneRequired                  = errors.New("customer phone is required")
)

type CustomerNotificationCommand struct {
	IdempotencyKey      string         `json:"idempotencyKey"`
	CustomerIDs         []string       `json:"customerIds"`
	NotificationChannel string         `json:"notificationChannel"`
	NotificationType    string         `json:"notificationType"`
	Body                map[string]any `json:"body"`
}

type Customer struct {
	ID      string
	Name    string
	Surname string
	Email   string
	Phone   string
	Note    string
}

type CustomerReader interface {
	GetCustomer(ctx context.Context, id string) (Customer, error)
}

type CustomerNotificationTemplateData struct {
	NotificationType    string
	NotificationChannel string
	Values              map[string]any
}

type CustomerNotificationTemplateRenderer interface {
	Render(ctx context.Context, data CustomerNotificationTemplateData) (string, error)
}

type CustomerNotificationDelivery struct {
	IdempotencyKey      string
	NotificationID      string
	CustomerID          string
	NotificationType    string
	NotificationChannel string
	CreatedAt           time.Time
}

type CustomerNotificationIdempotencyRepository interface {
	Exists(ctx context.Context, idempotencyKey string) (bool, error)
	Save(ctx context.Context, delivery CustomerNotificationDelivery) error
}

type CustomerNotificationService struct {
	customers     CustomerReader
	templates     CustomerNotificationTemplateRenderer
	idempotency   CustomerNotificationIdempotencyRepository
	notifications *NotificationService
	now           func() time.Time
}

func NewCustomerNotificationService(customers CustomerReader, templates CustomerNotificationTemplateRenderer, idempotency CustomerNotificationIdempotencyRepository, notifications *NotificationService) *CustomerNotificationService {
	return &CustomerNotificationService{customers: customers, templates: templates, idempotency: idempotency, notifications: notifications, now: time.Now}
}

func (service *CustomerNotificationService) Process(ctx context.Context, command CustomerNotificationCommand) error {
	if err := command.Validate(); err != nil {
		return err
	}
	if command.Body == nil {
		command.Body = map[string]any{}
	}
	for _, customerID := range command.CustomerIDs {
		if err := service.processCustomer(ctx, command, customerID); err != nil {
			return err
		}
	}
	return nil
}

func (service *CustomerNotificationService) processCustomer(ctx context.Context, command CustomerNotificationCommand, customerID string) error {
	key := command.CustomerIdempotencyKey(customerID)
	exists, err := service.idempotency.Exists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	customer, err := service.customers.GetCustomer(ctx, customerID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(customer.Phone) == "" {
		return fmt.Errorf("%w: %s", ErrCustomerPhoneRequired, customerID)
	}

	content, err := service.templates.Render(ctx, CustomerNotificationTemplateData{
		NotificationType:    command.NotificationType,
		NotificationChannel: command.NotificationChannel,
		Values:              command.TemplateValues(customer),
	})
	if err != nil {
		return err
	}

	notification, err := service.notifications.CreateNotification(ctx, "", content, domain.Channel{Type: domain.ChannelSMS, Phone: customer.Phone})
	if err != nil {
		return err
	}
	if err := service.notifications.SendNotification(ctx, notification.ID); err != nil {
		return err
	}
	return service.idempotency.Save(ctx, CustomerNotificationDelivery{
		IdempotencyKey:      key,
		NotificationID:      notification.ID,
		CustomerID:          customerID,
		NotificationType:    command.NotificationType,
		NotificationChannel: command.NotificationChannel,
		CreatedAt:           service.now().UTC(),
	})
}

func (command CustomerNotificationCommand) Validate() error {
	if len(command.CustomerIDs) == 0 {
		return errors.New("customerIds is required")
	}
	if strings.TrimSpace(command.NotificationType) == "" {
		return errors.New("notificationType is required")
	}
	if strings.ContainsAny(command.NotificationType, `/\\`) {
		return errors.New("notificationType cannot contain path separators")
	}
	if command.NotificationChannel != string(domain.ChannelSMS) {
		return fmt.Errorf("%w: %s", ErrUnsupportedCustomerNotificationChannel, command.NotificationChannel)
	}
	if command.Body == nil {
		command.Body = map[string]any{}
	}
	for _, customerID := range command.CustomerIDs {
		if strings.TrimSpace(customerID) == "" {
			return errors.New("customerIds cannot contain empty values")
		}
	}
	return nil
}

func (command CustomerNotificationCommand) TemplateValues(customer Customer) map[string]any {
	values := map[string]any{
		"id":                  customer.ID,
		"name":                customer.Name,
		"surname":             customer.Surname,
		"email":               customer.Email,
		"phone":               customer.Phone,
		"note":                customer.Note,
		"notificationType":    command.NotificationType,
		"notificationChannel": command.NotificationChannel,
	}
	for key, value := range command.Body {
		values[key] = value
	}
	return values
}

func (command CustomerNotificationCommand) CustomerIdempotencyKey(customerID string) string {
	base := strings.TrimSpace(command.IdempotencyKey)
	if base == "" {
		base = command.fallbackIdempotencyKey()
	}
	return fmt.Sprintf("%s:%s:%s:%s", base, customerID, command.NotificationChannel, command.NotificationType)
}

func (command CustomerNotificationCommand) fallbackIdempotencyKey() string {
	payload, err := json.Marshal(command)
	if err != nil {
		return uuid.NewString()
	}
	sum := sha256.Sum256(payload)
	return "fallback:" + hex.EncodeToString(sum[:])
}
