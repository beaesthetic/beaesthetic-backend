package http

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/token"
)

func TestIssueTokenRequiresInternalKey(t *testing.T) {
	handler := NewHandler(Config{InternalAPIKey: "secret"}, testIssuer(t), nil)
	request := httptest.NewRequest(http.MethodPost, "/v1/internal/tokens", strings.NewReader(`{"subject":"user_1","identity_type":"human","audience":"appointment","auth_method":"provider:xyz"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestIssueAndVerifyToken(t *testing.T) {
	handler := NewHandler(Config{InternalAPIKey: "secret"}, testIssuer(t), nil)
	issue := httptest.NewRequest(http.MethodPost, "/v1/internal/tokens", strings.NewReader(`{"subject":"user_1","identity_type":"human","audience":"appointment","permissions":["appointments:read"],"auth_method":"provider:xyz"}`))
	issue.Header.Set("X-Internal-API-Key", "secret")
	issued := httptest.NewRecorder()
	handler.ServeHTTP(issued, issue)
	if issued.Code != http.StatusCreated {
		t.Fatalf("issue status = %d: %s", issued.Code, issued.Body.String())
	}
	if !strings.Contains(issued.Body.String(), "v4.public.") {
		t.Fatalf("unexpected issue response: %s", issued.Body.String())
	}
}

func testIssuer(t *testing.T) *token.Issuer {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer, err := token.NewIssuer(token.Config{Issuer: "test", ActiveKeyID: "2026-01", PrivateKeyB64: base64.RawURLEncoding.EncodeToString(privateKey), TTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	return issuer
}
