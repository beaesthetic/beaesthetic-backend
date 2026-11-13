package customer

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	domainv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
)

type CustomerRegistry struct {
	api ClientWithResponsesInterface
}

func NewCustomerRegistry(baseURL string) (*CustomerRegistry, error) {
	api, err := NewClientWithResponses(baseURL)
	if err != nil {
		return nil, err
	}
	return &CustomerRegistry{api: api}, nil
}

func (r *CustomerRegistry) FindByCustomerID(ctx context.Context, customerID string) (*application.Customer, error) {
	response, err := r.api.GetInternalCustomerByIdWithResponse(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() == http.StatusNotFound {
		return nil, nil
	}
	if response.StatusCode() != http.StatusOK || response.JSON200 == nil {
		return nil, fmt.Errorf("customer service returned status %d for customer %s", response.StatusCode(), customerID)
	}
	customer := response.JSON200
	return &application.Customer{
		ID:          customer.Id,
		DisplayName: strings.TrimSpace(customer.Name + " " + customer.Surname),
		PhoneNumber: customer.Phone,
	}, nil
}

func (r *CustomerRegistry) ResolveCustomer(ctx context.Context, customerID string) (domainv2.CustomerRef, error) {
	customer, err := r.FindByCustomerID(ctx, customerID)
	if err != nil {
		return domainv2.CustomerRef{}, err
	}
	if customer == nil {
		return domainv2.CustomerRef{}, domainv2.ErrMissingRequiredData
	}
	return domainv2.NewCustomerRef(customer.ID, customer.DisplayName)
}
