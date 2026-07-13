package customer

import (
	"context"
	"fmt"
	"net/http"
	"time"

	customerapi "github.com/petretiandrea/beaesthetic-backend/notification/internal/api/customer"
	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
)

type Client struct {
	api customerapi.ClientWithResponsesInterface
}

func NewClient(baseURL string) (*Client, error) {
	api, err := customerapi.NewClientWithResponses(baseURL, customerapi.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}))
	if err != nil {
		return nil, err
	}
	return &Client{api: api}, nil
}

func (client *Client) GetCustomer(ctx context.Context, id string) (application.Customer, error) {
	resp, err := client.api.GetCustomerByIdWithResponse(ctx, id)
	if err != nil {
		return application.Customer{}, err
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return application.Customer{}, fmt.Errorf("customer service returned status %d for customer %s", resp.StatusCode(), id)
	}
	customer := resp.JSON200
	result := application.Customer{
		ID:      customer.Id,
		Name:    customer.Name,
		Surname: customer.Surname,
	}
	if customer.Email != nil {
		result.Email = string(*customer.Email)
	}
	if customer.Phone != nil {
		result.Phone = *customer.Phone
	}
	if customer.Note != nil {
		result.Note = *customer.Note
	}
	return result, nil
}
