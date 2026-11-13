package customer

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomerRegistryFindByCustomerIDUsesInternalEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/internal/customers/customer-1" {
			t.Fatalf("request = %s %s, want GET /internal/customers/customer-1", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"customer-1","name":"Ada","surname":"Lovelace","phone":"+393331234567","note":"must not be part of the contract"}`))
	}))
	defer server.Close()

	registry, err := NewCustomerRegistry(server.URL)
	if err != nil {
		t.Fatalf("NewCustomerRegistry() error = %v", err)
	}
	customer, err := registry.FindByCustomerID(context.Background(), "customer-1")
	if err != nil {
		t.Fatalf("FindByCustomerID() error = %v", err)
	}
	if customer == nil || customer.ID != "customer-1" || customer.DisplayName != "Ada Lovelace" || customer.PhoneNumber == nil || *customer.PhoneNumber != "+393331234567" {
		t.Fatalf("FindByCustomerID() = %#v, want customer data from internal endpoint", customer)
	}
}

func TestCustomerRegistryFindByCustomerIDReturnsNilForNotFound(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	registry, err := NewCustomerRegistry(server.URL)
	if err != nil {
		t.Fatalf("NewCustomerRegistry() error = %v", err)
	}

	customer, err := registry.FindByCustomerID(context.Background(), "missing")
	if err != nil || customer != nil {
		t.Fatalf("FindByCustomerID() = (%#v, %v), want (nil, nil)", customer, err)
	}
}

func TestCustomerRegistryFindByCustomerIDReturnsTransportError(t *testing.T) {
	registry, err := NewCustomerRegistry("https://customer.example")
	if err != nil {
		t.Fatalf("NewCustomerRegistry() error = %v", err)
	}
	expected := errors.New("connection refused")
	api, err := NewClientWithResponses("https://customer.example", WithHTTPClient(&http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, expected
	})}))
	if err != nil {
		t.Fatalf("NewClientWithResponses() error = %v", err)
	}
	registry.api = api

	_, err = registry.FindByCustomerID(context.Background(), "customer-1")
	if !errors.Is(err, expected) {
		t.Fatalf("FindByCustomerID() error = %v, want transport error %v", err, expected)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }
