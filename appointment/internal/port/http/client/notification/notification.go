package notification

import (
	"context"
	"fmt"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type NotificationClient struct {
	client *ClientWithResponses
}

func NewNotificationClient(baseURL string) (*NotificationClient, error) {
	client, err := NewClientWithResponses(baseURL)
	if err != nil {
		return nil, err
	}
	return &NotificationClient{client: client}, nil
}

func (c *NotificationClient) Raw() *ClientWithResponses {
	return c.client
}

func (c *NotificationClient) Send(ctx context.Context, title, content, phoneNumber string) (string, error) {
	response, err := c.client.CreateNotificationWithResponse(ctx, CreateNotificationJSONRequestBody{
		Title:   title,
		Content: content,
		Channel: NotificationChannel{
			Type:  Sms,
			Phone: &phoneNumber,
		},
	})
	if err != nil {
		return "", err
	}
	if response.JSON200 == nil || response.JSON200.NotificationId == nil {
		return "", fmt.Errorf("notification service returned empty notification id")
	}
	return openapi_types.UUID(*response.JSON200.NotificationId).String(), nil
}
