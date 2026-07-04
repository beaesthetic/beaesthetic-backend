package httpclient

import notificationclient "github.com/petretiandrea/beaesthetic-backend/appointment/internal/port/http/notificationclient"

type NotificationClient struct {
	client *notificationclient.ClientWithResponses
}

func NewNotificationClient(baseURL string) (*NotificationClient, error) {
	client, err := notificationclient.NewClientWithResponses(baseURL)
	if err != nil {
		return nil, err
	}
	return &NotificationClient{client: client}, nil
}

func (c *NotificationClient) Raw() *notificationclient.ClientWithResponses {
	return c.client
}
