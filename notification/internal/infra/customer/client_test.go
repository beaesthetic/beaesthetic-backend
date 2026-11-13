package customer

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	customerapi "github.com/petretiandrea/beaesthetic-backend/notification/internal/api/customer"
)

func TestClientGetCustomerUsesInternalEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal/customers/customer-1" {
			t.Fatalf("request = %s %s, want GET /internal/customers/customer-1", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"customer-1","name":"Ada","surname":"Lovelace","email":"ada@example.com","phone":"+393331234567","note":"must not be part of the contract"}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	customer, err := client.GetCustomer(context.Background(), "customer-1")
	if err != nil {
		t.Fatalf("GetCustomer() error = %v", err)
	}
	if customer.ID != "customer-1" || customer.Name != "Ada" || customer.Surname != "Lovelace" || customer.Email != "ada@example.com" || customer.Phone != "+393331234567" || customer.Note != "" {
		t.Fatalf("GetCustomer() = %#v, want internal customer without note", customer)
	}
}

func TestClientGetCustomerReturnsNotFoundError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.GetCustomer(context.Background(), "missing")
	if err == nil || !strings.Contains(err.Error(), "status 404") {
		t.Fatalf("GetCustomer() error = %v, want 404 error", err)
	}
}

func TestClientGetCustomerReturnsTransportError(t *testing.T) {
	client, err := NewClient("https://customer.example")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	expected := errors.New("connection refused")
	api, err := customerapi.NewClientWithResponses("https://customer.example", customerapi.WithHTTPClient(roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, expected
	}))
	if err != nil {
		t.Fatalf("NewClientWithResponses() error = %v", err)
	}
	client.api = api

	_, err = client.GetCustomer(context.Background(), "customer-1")
	if !errors.Is(err, expected) {
		t.Fatalf("GetCustomer() error = %v, want transport error %v", err, expected)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }
