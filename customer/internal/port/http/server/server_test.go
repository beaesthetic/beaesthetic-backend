package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	customerapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/customer"
)

func TestNewRegistersCompatibleRoutes(t *testing.T) {
	t.Parallel()

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("New panicked while registering routes: %v", recovered)
		}
	}()

	_ = New(&HttpHandlers{Customer: &Server{}, Fidelity: &Server{}, Wallet: &Server{}}, nil)
}

func TestNewNormalizesTrailingSlashWithoutRedirect(t *testing.T) {
	t.Parallel()

	router := New(&HttpHandlers{Customer: &Server{}, Fidelity: &Server{}, Wallet: &Server{}}, nil)
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
