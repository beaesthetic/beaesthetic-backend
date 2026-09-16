package oauth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/membership"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/token"
)

func TestExchangeResolvesAuthorizationFromMembership(t *testing.T) {
	issuer := testIssuer(t)
	repository := memoryRepository{user: membership.User{ID: "usr_1"}, record: membership.Membership{ID: "mem_1", UserID: "usr_1", OrganizationID: "org_1", Active: true, Roles: []string{"owner"}, Permissions: []string{"users:read", "users:write"}}}
	service := NewExchangeService(issuer, []Authenticator{testAuthenticator{}}, []Authorizer{NewMembershipAuthorizer(repository)})
	_, claims, err := service.Exchange(context.Background(), Request{GrantType: TokenExchangeGrant, SubjectTokenType: JWTSubjectTokenType, Authenticator: "firebase", Authorizer: "membership", SubjectToken: "external-token", Audience: "appointment", OrganizationID: "org_1"})
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "usr_1" || claims.OrganizationID != "org_1" || claims.MembershipID != "mem_1" {
		t.Fatalf("unexpected identity claims: %+v", claims)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "owner" || len(claims.Permissions) != 2 {
		t.Fatalf("unexpected authorization claims: %+v", claims)
	}
}

type testAuthenticator struct{}

func (testAuthenticator) Name() string { return "firebase" }
func (testAuthenticator) Authenticate(context.Context, string) (ExternalIdentity, error) {
	return ExternalIdentity{Provider: "firebase", Subject: "firebase-uid", Email: "user@example.com"}, nil
}

type memoryRepository struct {
	user   membership.User
	record membership.Membership
}

func (r memoryRepository) FindOrCreateUser(context.Context, string, string, string) (membership.User, error) {
	return r.user, nil
}
func (r memoryRepository) FindActiveMembership(context.Context, string, string) (membership.Membership, error) {
	return r.record, nil
}

func testIssuer(t *testing.T) *token.Issuer {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer, err := token.NewIssuer(token.Config{Issuer: "identity", ActiveKeyID: "key-1", PrivateKeyB64: base64.RawURLEncoding.EncodeToString(privateKey), TTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	return issuer
}
