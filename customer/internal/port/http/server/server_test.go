package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	customerapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/customer"
	fidelityapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/fidelity"
	walletapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/wallet"
)

func TestNewRegistersCompatibleRoutes(t *testing.T) {
	t.Parallel()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("New panicked while registering routes: %v", recovered)
		}
	}()

	router := newRouter(&HttpHandlers{Customer: &Server{}, Fidelity: &Server{}, Wallet: &Server{}}, nil, nil)
	for _, route := range router.Routes() {
		if route.Method == http.MethodGet && route.Path == "/internal/customers/:customerId" {
			return
		}
	}
	t.Fatal("internal customer lookup route is not registered")
}

func TestNewNormalizesTrailingSlashWithoutRedirect(t *testing.T) {
	t.Parallel()

	router := newRouter(&HttpHandlers{Customer: &Server{}, Fidelity: &Server{}, Wallet: &Server{}}, nil, nil)
	request := httptest.NewRequest(http.MethodPost, "/admin/wallets/giftCard/", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code == http.StatusTemporaryRedirect || response.Code == http.StatusMovedPermanently {
		t.Fatalf("expected no redirect, got status %d", response.Code)
	}
}

func TestCustomerCreateAcceptsEmptyEmail(t *testing.T) {
	t.Parallel()

	var request customerapi.CustomerCreate
	if err := json.Unmarshal([]byte(`{"name":"ciao","surname":"aa","email":"","phone":"+391234344356546","note":""}`), &request); err != nil {
		t.Fatalf("CustomerCreate with empty email should decode: %v", err)
	}
	if request.Email == nil || *request.Email != "" {
		t.Fatalf("Email = %#v, want empty string pointer", request.Email)
	}
}

// Manual registration must stay in sync with the generated contracts.
func TestRegisteredRoutesMatchContracts(t *testing.T) {
	actual := newRouter(&HttpHandlers{Customer: &Server{}, Fidelity: &Server{}, Wallet: &Server{}}, nil, nil)
	expected := gin.New()
	customerapi.RegisterHandlers(expected, nil)
	fidelityapi.RegisterHandlers(expected, nil)
	walletapi.RegisterHandlers(expected, nil)
	noop := func(*gin.Context) {}
	expected.GET("/admin/fidelity-cards/", noop)
	expected.POST("/admin/fidelity-cards/", noop)
	expected.GET("/admin/wallets/", noop)
	expected.POST("/admin/wallets/giftCard/", noop)
	expected.GET("/health", noop)
	routes := map[string]bool{}
	for _, route := range expected.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range actual.Routes() {
		key := route.Method + " " + route.Path
		if !routes[key] {
			t.Errorf("unexpected route %s", key)
		}
		delete(routes, key)
	}
	for key := range routes {
		t.Errorf("missing route %s", key)
	}
}
