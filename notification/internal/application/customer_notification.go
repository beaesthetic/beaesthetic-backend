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

type CustomerNotificationSMSDispatcher interface {
	Send(ctx context.Context, messageID, phone, content string) (string, error)
}

type CustomerNotificationRecord struct {
	ID                  string
	IdempotencyKey      string
	CorrelationKey      string
	CustomerID          string
	NotificationType    string
	NotificationChannel string
	TemplateValues      map[string]any
	Status              string
	CreatedAt           time.Time
	SentAt              time.Time
}

type SMSGatewayDispatch struct {
	ID                     string
	CustomerNotificationID string
	SMSGatewayMessageID    string
	CreatedAt              time.Time
}

type CustomerNotificationRepository interface {
	Exists(ctx context.Context, idempotencyKey string) (bool, error)
	CreatePending(ctx context.Context, notification CustomerNotificationRecord) (bool, error)
	SaveSMSGatewayDispatch(ctx context.Context, dispatch SMSGatewayDispatch) error
	MarkSentBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, sentAt time.Time) (bool, error)
	MarkFailedBySMSGatewayMessageID(ctx context.Context, smsGatewayMessageID string, failedAt time.Time) (bool, error)
}

type CustomerNotificationService struct {
	customers     CustomerReader
	templates     CustomerNotificationTemplateRenderer
	repository    CustomerNotificationRepository
	smsDispatcher CustomerNotificationSMSDispatcher
	now           func() time.Time
}

func NewCustomerNotificationService(customers CustomerReader, templates CustomerNotificationTemplateRenderer, repository CustomerNotificationRepository, smsDispatcher CustomerNotificationSMSDispatcher) *CustomerNotificationService {
	return &CustomerNotificationService{customers: customers, templates: templates, repository: repository, smsDispatcher: smsDispatcher, now: time.Now}
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

func (service *CustomerNotificationService) ConfirmSMSGatewayMessageSent(ctx context.Context, smsGatewayMessageID string) (bool, error) {
	return service.repository.MarkSentBySMSGatewayMessageID(ctx, smsGatewayMessageID, service.now().UTC())
}

func (service *CustomerNotificationService) MarkSMSGatewayMessageFailed(ctx context.Context, smsGatewayMessageID string) (bool, error) {
	return service.repository.MarkFailedBySMSGatewayMessageID(ctx, smsGatewayMessageID, service.now().UTC())
}

func (service *CustomerNotificationService) processCustomer(ctx context.Context, command CustomerNotificationCommand, customerID string) error {
	correlationKey := command.CorrelationKey()
	key := command.CustomerIdempotencyKey(customerID)
	exists, err := service.repository.Exists(ctx, key)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	newNotificationID := uuid.NewString()
	now := service.now().UTC()
	created, err := service.repository.CreatePending(ctx, CustomerNotificationRecord{
		ID:                  newNotificationID,
		IdempotencyKey:      key,
		CorrelationKey:      correlationKey,
		CustomerID:          customerID,
		NotificationType:    command.NotificationType,
		NotificationChannel: command.NotificationChannel,
		TemplateValues:      command.Body,
		Status:              "pending",
		CreatedAt:           now,
	})
	if err != nil {
		return err
	}
	if !created {
		return nil
	}

	customer, err := service.customers.GetCustomer(ctx, customerID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(customer.Phone) == "" {
		return fmt.Errorf("%w: %s", ErrCustomerPhoneRequired, customerID)
	}
	templateValues := command.TemplateValues(customer)
	content, err := service.templates.Render(ctx, CustomerNotificationTemplateData{
		NotificationType:    command.NotificationType,
		NotificationChannel: command.NotificationChannel,
		Values:              templateValues,
	})
	if err != nil {
		return err
	}

	smsGatewayMessageID, err := service.smsDispatcher.Send(ctx, newNotificationID, customer.Phone, content)
	if err != nil {
		return err
	}
	if smsGatewayMessageID == "" {
		return nil
	}
	return service.repository.SaveSMSGatewayDispatch(ctx, SMSGatewayDispatch{
		ID:                     uuid.NewString(),
		CustomerNotificationID: newNotificationID,
		SMSGatewayMessageID:    smsGatewayMessageID,
		CreatedAt:              service.now().UTC(),
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
	if command.NotificationChannel != "sms" {
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
	return fmt.Sprintf("%s:%s:%s:%s", command.CorrelationKey(), customerID, command.NotificationChannel, command.NotificationType)
}

func (command CustomerNotificationCommand) CorrelationKey() string {
	base := strings.TrimSpace(command.IdempotencyKey)
	if base != "" {
		return base
	}
	return command.fallbackIdempotencyKey()
}

func (command CustomerNotificationCommand) fallbackIdempotencyKey() string {
	payload, err := json.Marshal(command)
	if err != nil {
		return uuid.NewString()
	}
	sum := sha256.Sum256(payload)
	return "fallback:" + hex.EncodeToString(sum[:])
}
