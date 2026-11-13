package customer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/application"
	domainv2 "github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain/v2"
)

type CustomerRegistry struct {
	baseURL    *url.URL
	httpClient *http.Client
}

func NewCustomerRegistry(baseURL string) (*CustomerRegistry, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("customer service URL must be absolute: %q", baseURL)
	}
	parsedURL.Path = strings.TrimRight(parsedURL.Path, "/") + "/"
	return &CustomerRegistry{baseURL: parsedURL, httpClient: http.DefaultClient}, nil
}

func (r *CustomerRegistry) FindByCustomerID(ctx context.Context, customerID string) (*application.Customer, error) {
	endpoint := r.baseURL.ResolveReference(&url.URL{Path: "internal/customers/" + url.PathEscape(customerID)})
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := r.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("customer service returned status %d for customer %s", response.StatusCode, customerID)
	}
	var customer struct {
		ID      string  `json:"id"`
		Name    string  `json:"name"`
		Surname string  `json:"surname"`
		Phone   *string `json:"phone"`
	}
	if err := json.NewDecoder(response.Body).Decode(&customer); err != nil {
		return nil, fmt.Errorf("decode customer service response: %w", err)
	}
	return &application.Customer{
		ID:          customer.ID,
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
