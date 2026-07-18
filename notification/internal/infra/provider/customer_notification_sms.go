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
)

type CustomerNotificationSMSDispatcher struct {
	client *http.Client
	config config.SMSGatewayConfig
}

func NewCustomerNotificationSMSDispatcher(config config.SMSGatewayConfig) *CustomerNotificationSMSDispatcher {
	return &CustomerNotificationSMSDispatcher{
		client: &http.Client{Timeout: 10 * time.Second},
		config: config,
	}
}

func (dispatcher *CustomerNotificationSMSDispatcher) Send(ctx context.Context, messageID, phone, content string) (string, error) {
	body := sendSMSRequest{
		Content: content,
		From:    dispatcher.config.FromNumber,
		To:      phone,
		Metadata: map[string]string{
			NotificationIDMetadata: messageID,
		},
	}
	if dispatcher.config.WebhookURL != "" {
		body.Webhook = &webhookConfig{URL: dispatcher.config.WebhookURL}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	url := strings.TrimRight(dispatcher.config.URL, "/") + "/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", messageID)
	if dispatcher.config.APIKey != "" {
		req.Header.Set("Api-Key", dispatcher.config.APIKey)
	}
	resp, err := dispatcher.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("sms gateway returned status %d", resp.StatusCode)
	}
	var response smsEntityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}
	return response.ID, nil
}
