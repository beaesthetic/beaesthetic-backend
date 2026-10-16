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
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/user"
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
	issue := httptest.NewRequest(http.MethodPost, "/v1/internal/tokens", strings.NewReader(`{"subject":"user_1","identityType":"IDENTITY_TYPE_HUMAN","permissions":["identity.organizations.read"],"authMethod":"provider:xyz"}`))
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
	request := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(`{"grantType":"urn:ietf:params:oauth:grant-type:token-exchange","subjectAssertion":{"authenticator":"test","token":"external-token","tokenType":"urn:ietf:params:oauth:token-type:jwt"},"authorizer":"test","organizationId":"org_1"}`))
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

func TestUserScopeEndpointsDoNotRequirePermissions(t *testing.T) {
	issuer := testIssuer(t)
	handler := New(Config{}, issuer, nil, httpMembershipRepository{})
	token := signedToken(t, issuer, token.IssueRequest{Subject: "user_1", IdentityType: token.IdentityHuman, Scope: token.ScopeUser})

	for _, path := range []string{"/v1/me/organizations", "/v1/organizations/org_1/authorization"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d: %s", path, response.Code, response.Body.String())
		}
	}
}

func TestMembershipManagementPermissions(t *testing.T) {
	issuer := testIssuer(t)
	handler := New(Config{}, issuer, nil, httpMembershipRepository{})
	readToken := signedToken(t, issuer, token.IssueRequest{Subject: "user_1", IdentityType: token.IdentityHuman, Scope: token.ScopeMembership, OrganizationID: "org_1", MembershipID: "membership_1", Permissions: []string{"identity.memberships.read"}})
	writeToken := signedToken(t, issuer, token.IssueRequest{Subject: "user_1", IdentityType: token.IdentityHuman, Scope: token.ScopeMembership, OrganizationID: "org_1", MembershipID: "membership_1", Permissions: []string{"identity.memberships.write"}})

	list := httptest.NewRequest(http.MethodGet, "/v1/organizations/org_1/members", nil)
	list.Header.Set("Authorization", "Bearer "+readToken)
	listed := httptest.NewRecorder()
	handler.ServeHTTP(listed, list)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status = %d: %s", listed.Code, listed.Body.String())
	}

	denied := httptest.NewRequest(http.MethodPost, "/v1/organizations/org_1/members", strings.NewReader(`{"email":"member@example.com","roles":["staff"]}`))
	denied.Header.Set("Authorization", "Bearer "+readToken)
	deniedResponse := httptest.NewRecorder()
	handler.ServeHTTP(deniedResponse, denied)
	if deniedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("read token post status = %d: %s", deniedResponse.Code, deniedResponse.Body.String())
	}

	add := httptest.NewRequest(http.MethodPost, "/v1/organizations/org_1/members", strings.NewReader(`{"email":"member@example.com","roles":["staff"]}`))
	add.Header.Set("Authorization", "Bearer "+writeToken)
	added := httptest.NewRecorder()
	handler.ServeHTTP(added, add)
	if added.Code != http.StatusCreated {
		t.Fatalf("write token post status = %d: %s", added.Code, added.Body.String())
	}
}

func TestNonOwnerCannotAssignOrRemoveOwner(t *testing.T) {
	issuer := testIssuer(t)
	handler := New(Config{}, issuer, nil, httpMembershipRepository{})
	accessToken := signedToken(t, issuer, token.IssueRequest{Subject: "admin_1", IdentityType: token.IdentityHuman, Scope: token.ScopeMembership, OrganizationID: "org_1", MembershipID: "membership_admin", Roles: []string{"admin"}, Permissions: []string{"identity.memberships.write"}})

	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodPost, "/v1/organizations/org_1/members", strings.NewReader(`{"email":"owner@example.com","roles":["owner"]}`)),
		httptest.NewRequest(http.MethodPut, "/v1/organizations/org_1/members/user_1/roles", strings.NewReader(`{"roles":["owner"]}`)),
		httptest.NewRequest(http.MethodDelete, "/v1/organizations/org_1/members/user_1", nil),
	} {
		request.Header.Set("Authorization", "Bearer "+accessToken)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Fatalf("%s %s status = %d: %s", request.Method, request.URL.Path, response.Code, response.Body.String())
		}
	}
}

func TestRemoveMember(t *testing.T) {
	issuer := testIssuer(t)
	repository := &removingMembershipRepository{}
	handler := New(Config{}, issuer, nil, repository)
	accessToken := signedToken(t, issuer, token.IssueRequest{
		Subject:        "owner_1",
		IdentityType:   token.IdentityHuman,
		Scope:          token.ScopeMembership,
		OrganizationID: "org_1",
		MembershipID:   "membership_owner",
		Roles:          []string{"owner"},
		Permissions:    []string{"identity.memberships.write"},
	})
	request := httptest.NewRequest(http.MethodDelete, "/v1/organizations/org_1/members/user_1", nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	if !repository.removed {
		t.Fatal("expected membership removal")
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
	return token.IssueRequest{Subject: "user_1", IdentityType: token.IdentityHuman, Scope: token.ScopeUser, AuthMethod: "test"}, nil
}

type httpMembershipRepository struct{}

type removingMembershipRepository struct {
	httpMembershipRepository
	removed bool
}

func (r *removingMembershipRepository) RemoveMember(context.Context, string, string) error {
	r.removed = true
	return nil
}

func (httpMembershipRepository) FindOrCreate(context.Context, string, string, string) (user.User, error) {
	return user.User{}, nil
}
func (httpMembershipRepository) FindActiveMembership(context.Context, string, string) (membership.Membership, error) {
	return membership.Membership{ID: "member_1", UserID: "user_1", OrganizationID: "org_1", Active: true, Roles: []string{"owner"}, Permissions: []string{"appointment.read"}}, nil
}
func (httpMembershipRepository) ListOrganizations(context.Context, string) ([]membership.Organization, error) {
	return []membership.Organization{{ID: "org_1", Name: "Organization", MembershipID: "member_1", Roles: []string{"owner"}}}, nil
}
func (httpMembershipRepository) ListMembers(context.Context, string) ([]membership.Member, error) {
	return []membership.Member{{UserID: "user_1", Email: "owner@example.com", MembershipID: "member_1", Active: true, Roles: []string{"owner"}}}, nil
}
func (httpMembershipRepository) AddMember(_ context.Context, _ string, email string, roles []string) (membership.Member, error) {
	return membership.Member{UserID: "member_2", Email: email, MembershipID: "membership_2", Active: true, Roles: roles}, nil
}
func (httpMembershipRepository) ReplaceMemberRoles(_ context.Context, _ string, userID string, roles []string) (membership.Member, error) {
	return membership.Member{UserID: userID, MembershipID: "member_1", Active: true, Roles: roles}, nil
}
func (httpMembershipRepository) RemoveMember(context.Context, string, string) error { return nil }
func (httpMembershipRepository) ListGlobalRoles(context.Context) ([]string, error) {
	return []string{"owner"}, nil
}

func signedToken(t *testing.T, issuer *token.Issuer, request token.IssueRequest) string {
	t.Helper()
	if request.AuthMethod == "" {
		request.AuthMethod = "test"
	}
	signed, _, err := issuer.Issue(request)
	if err != nil {
		t.Fatal(err)
	}
	return signed
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
