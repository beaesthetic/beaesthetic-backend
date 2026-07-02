package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/domain"
	"go.uber.org/zap"
)

func TestCreateNotification(t *testing.T) {
	repo := &memoryRepo{}
	service := application.NewNotificationService(repo, fakeProvider{})
	router := NewRouter(NewServer(service, zap.NewNop()))

	body := []byte(`{"title":"Welcome","content":"Hello","channel":{"type":"sms","phone":"+393331234567"}}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var payload struct {
		NotificationID string `json:"notificationId"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.NotificationID == "" {
		t.Fatal("notificationId should be returned")
	}
}

func TestWebhookRequiresNotificationMetadata(t *testing.T) {
	service := application.NewNotificationService(&memoryRepo{}, fakeProvider{})
	router := NewRouter(NewServer(service, zap.NewNop()))

	body := []byte(`{"eventType":"message.deliver.succeeded","metadata":{}}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/sms-webhook/notify", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

type memoryRepo struct {
	notifications map[string]*domain.Notification
}

func (repo *memoryRepo) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	if repo.notifications == nil {
		return nil, nil
	}
	return repo.notifications[id], nil
}

func (repo *memoryRepo) Save(ctx context.Context, notification *domain.Notification) error {
	if repo.notifications == nil {
		repo.notifications = make(map[string]*domain.Notification)
	}
	repo.notifications[notification.ID] = notification
	notification.PullEvents()
	return nil
}

type fakeProvider struct{}

func (fakeProvider) Supports(notification domain.Notification) bool {
	return notification.Channel.Type == domain.ChannelSMS
}

func (fakeProvider) Send(ctx context.Context, notification domain.Notification) (domain.ChannelMetadata, error) {
	return domain.ChannelMetadata{ProviderResourceID: "provider-id"}, nil
}
