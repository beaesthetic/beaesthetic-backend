package customer

import (
	"context"
	"strings"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
)

type CustomerRegistry struct {
	client *ClientWithResponses
}

func NewCustomerRegistry(baseURL string) (*CustomerRegistry, error) {
	client, err := NewClientWithResponses(baseURL)
	if err != nil {
		return nil, err
	}
	return &CustomerRegistry{client: client}, nil
}

func (r *CustomerRegistry) FindByCustomerID(ctx context.Context, customerID string) (*application.Customer, error) {
	response, err := r.client.GetCustomerByIdWithResponse(ctx, customerID)
	if err != nil || response.JSON200 == nil {
		return nil, err
	}
	customer := response.JSON200
	return &application.Customer{
		ID:          customer.Id,
		DisplayName: strings.TrimSpace(customer.Name + " " + customer.Surname),
		PhoneNumber: customer.Phone,
	}, nil
}
