package server

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/membership"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/token"
)

func TestIssueTokenRequiresInternalKey(t *testing.T) {
	handler := New(Config{InternalAPIKey: "secret"}, testIssuer(t), nil, nil)
	request := httptest.NewRequest(http.MethodPost, "/v1/internal/tokens", strings.NewReader(`{"subject":"user_1","identity_type":"human","audience":"appointment","auth_method":"provider:xyz"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestIssueAndVerifyToken(t *testing.T) {
	handler := New(Config{InternalAPIKey: "secret"}, testIssuer(t), nil, nil)
	issue := httptest.NewRequest(http.MethodPost, "/v1/internal/tokens", strings.NewReader(`{"subject":"user_1","identityType":"IDENTITY_TYPE_HUMAN","audience":"appointment","permissions":["appointments:read"],"authMethod":"provider:xyz"}`))
	issue.Header.Set("X-Internal-API-Key", "secret")
	issued := httptest.NewRecorder()
	handler.ServeHTTP(issued, issue)
	if issued.Code != http.StatusCreated {
		t.Fatalf("issue status = %d: %s", issued.Code, issued.Body.String())
	}
	if !strings.Contains(issued.Body.String(), `"accessToken":"v4.public.`) {
		t.Fatalf("unexpected issue response: %s", issued.Body.String())
	}
}

func TestExchangeTokenAcceptsProtoJSONShape(t *testing.T) {
	exchange := application.NewExchangeService(testIssuer(t), []application.Authenticator{httpAuthenticator{}}, []application.Authorizer{httpAuthorizer{}})
	handler := New(Config{InternalAPIKey: "secret"}, testIssuer(t), exchange, nil)
	request := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(`{"grantType":"urn:ietf:params:oauth:grant-type:token-exchange","subjectAssertion":{"authenticator":"test","token":"external-token","tokenType":"urn:ietf:params:oauth:token-type:jwt"},"authorizer":"test","audience":"appointment","organizationId":"org_1"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"accessToken":"v4.public.`) {
		t.Fatalf("exchange response = %d: %s", response.Code, response.Body.String())
	}
}

func TestEffectiveAuthorizationRequiresInternalKey(t *testing.T) {
	handler := New(Config{InternalAPIKey: "secret"}, testIssuer(t), nil, httpMembershipRepository{})
	request := httptest.NewRequest(http.MethodGet, "/v1/internal/organizations/org_1/users/user_1/authorization", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestEffectiveAuthorization(t *testing.T) {
	handler := New(Config{InternalAPIKey: "secret"}, testIssuer(t), nil, httpMembershipRepository{})
	request := httptest.NewRequest(http.MethodGet, "/v1/internal/organizations/org_1/users/user_1/authorization", nil)
	request.Header.Set("X-Internal-API-Key", "secret")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"userId":"user_1"`) {
		t.Fatalf("authorization response = %d: %s", response.Code, response.Body.String())
	}
}

type httpAuthenticator struct{}

func (httpAuthenticator) Name() string { return "test" }
func (httpAuthenticator) Authenticate(context.Context, string) (application.ExternalIdentity, error) {
	return application.ExternalIdentity{Provider: "test", Subject: "subject"}, nil
}

type httpAuthorizer struct{}

func (httpAuthorizer) Name() string { return "test" }
func (httpAuthorizer) Authorize(context.Context, application.ExternalIdentity, string) (token.IssueRequest, error) {
	return token.IssueRequest{Subject: "user_1", IdentityType: token.IdentityHuman, AuthMethod: "test"}, nil
}

type httpMembershipRepository struct{}

func (httpMembershipRepository) FindOrCreateUser(context.Context, string, string, string) (membership.User, error) {
	return membership.User{}, nil
}
func (httpMembershipRepository) FindActiveMembership(context.Context, string, string) (membership.Membership, error) {
	return membership.Membership{ID: "member_1", UserID: "user_1", OrganizationID: "org_1", Active: true, Roles: []string{"admin"}, Permissions: []string{"appointments:read"}}, nil
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
