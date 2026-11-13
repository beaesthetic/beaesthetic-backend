package customer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/notification/internal/application"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

func NewClient(baseURL string) (*Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, fmt.Errorf("customer service URL must be absolute: %q", baseURL)
	}
	parsedURL.Path = strings.TrimRight(parsedURL.Path, "/") + "/"
	return &Client{baseURL: parsedURL, httpClient: &http.Client{Timeout: 10 * time.Second}}, nil
}

func (client *Client) GetCustomer(ctx context.Context, id string) (application.Customer, error) {
	endpoint := client.baseURL.ResolveReference(&url.URL{Path: "internal/customers/" + url.PathEscape(id)})
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return application.Customer{}, err
	}
	resp, err := client.httpClient.Do(request)
	if err != nil {
		return application.Customer{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return application.Customer{}, fmt.Errorf("customer service returned status %d for customer %s", resp.StatusCode, id)
	}
	var customer struct {
		ID      string  `json:"id"`
		Name    string  `json:"name"`
		Surname string  `json:"surname"`
		Email   *string `json:"email"`
		Phone   *string `json:"phone"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&customer); err != nil {
		return application.Customer{}, fmt.Errorf("decode customer service response: %w", err)
	}
	result := application.Customer{
		ID:      customer.ID,
		Name:    customer.Name,
		Surname: customer.Surname,
	}
	if customer.Email != nil {
		result.Email = *customer.Email
	}
	if customer.Phone != nil {
		result.Phone = *customer.Phone
	}
	return result, nil
}
