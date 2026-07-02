package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/config"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/domain"
)

const NotificationIDMetadata = "notificationId"

type SMSProvider struct {
	client *http.Client
	config config.SMSGatewayConfig
}

func NewSMSProvider(config config.SMSGatewayConfig) *SMSProvider {
	return &SMSProvider{
		client: &http.Client{Timeout: 10 * time.Second},
		config: config,
	}
}

func (provider *SMSProvider) Supports(notification domain.Notification) bool {
	return notification.Channel.Type == domain.ChannelSMS
}

func (provider *SMSProvider) Send(ctx context.Context, notification domain.Notification) (domain.ChannelMetadata, error) {
	body := sendSMSRequest{
		Content: notification.Content,
		From:    provider.config.FromNumber,
		To:      notification.Channel.Phone,
		Metadata: map[string]string{
			NotificationIDMetadata: notification.ID,
		},
	}
	if provider.config.WebhookURL != "" {
		body.Webhook = &webhookConfig{URL: provider.config.WebhookURL}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return domain.ChannelMetadata{}, err
	}
	url := strings.TrimRight(provider.config.URL, "/") + "/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return domain.ChannelMetadata{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", notification.ID)
	if provider.config.APIKey != "" {
		req.Header.Set("Api-Key", provider.config.APIKey)
	}
	resp, err := provider.client.Do(req)
	if err != nil {
		return domain.ChannelMetadata{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.ChannelMetadata{}, fmt.Errorf("sms gateway returned status %d", resp.StatusCode)
	}
	var response smsEntityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return domain.ChannelMetadata{}, err
	}
	return domain.ChannelMetadata{ProviderResourceID: response.ID}, nil
}

type sendSMSRequest struct {
	Content  string            `json:"content"`
	From     string            `json:"from"`
	To       string            `json:"to"`
	Webhook  *webhookConfig    `json:"webhook,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type webhookConfig struct {
	URL string `json:"url"`
}

type smsEntityResponse struct {
	ID string `json:"id"`
}
