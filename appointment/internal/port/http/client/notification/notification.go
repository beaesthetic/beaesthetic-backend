package notification

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
