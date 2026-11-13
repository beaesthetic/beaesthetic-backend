package token_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"reflect"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/token"
)

func TestIssuerCoreAuthCompatibility(t *testing.T) {
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	issuer, err := token.NewIssuer(token.Config{Issuer: "identity", ActiveKeyID: "active", PrivateKeyB64: base64.RawURLEncoding.EncodeToString(private), TTL: time.Minute, Clock: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]string{}
	for _, key := range issuer.PublicKeys() {
		keys[key.KeyID] = key.PublicKeyB64
	}
	verifier, err := auth.NewVerifier(auth.VerifierConfig{Issuer: "identity", PublicKeys: keys, Clock: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	for _, identityType := range []auth.IdentityType{auth.IdentityHuman, auth.IdentityService, auth.IdentityIntegration} {
		for _, scope := range []auth.Scope{auth.ScopeUser, auth.ScopeMembership} {
			t.Run(string(identityType)+"/"+string(scope), func(t *testing.T) {
				request := token.IssueRequest{Subject: "person-1", IdentityType: identityType, Scope: scope, AuthMethod: "test"}
				if scope == auth.ScopeMembership {
					request.OrganizationID = "org-1"
					request.MembershipID = "member-1"
					request.Roles = []string{"staff", "staff"}
					request.Permissions = []string{"identity.memberships.read", " identity.memberships.read "}
				}
				signed, issued, err := issuer.Issue(request)
				if err != nil {
					t.Fatal(err)
				}
				verified, err := verifier.Verify(signed)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(issued, verified) {
					t.Fatalf("issued %+v != verified %+v", issued, verified)
				}
				// Inspect the signed payload too: JSON omission rules on Claims must not
				// remove fields Identity has historically emitted for either scope.
				public, err := paseto.NewV4AsymmetricPublicKeyFromBytes(private.Public().(ed25519.PublicKey))
				if err != nil {
					t.Fatal(err)
				}
				payload, err := paseto.NewParserWithoutExpiryCheck().ParseV4Public(public, signed, nil)
				if err != nil {
					t.Fatal(err)
				}
				for _, field := range []string{"roles", "permissions"} {
					var values []string
					if err := payload.Get(field, &values); err != nil || values == nil {
						t.Fatalf("%s must be an array: %v", field, err)
					}
				}
				for field, want := range map[string]string{"organization_id": request.OrganizationID, "membership_id": request.MembershipID} {
					var got string
					if err := payload.Get(field, &got); err != nil || got != want {
						t.Fatalf("%s = %q, want %q: %v", field, got, want, err)
					}
				}
				if _, err := verifier.Verify(signed + "x"); err == nil {
					t.Fatal("tampered token accepted")
				}
				expired, err := auth.NewVerifier(auth.VerifierConfig{Issuer: "identity", PublicKeys: keys, Clock: func() time.Time { return now.Add(time.Minute) }})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := expired.Verify(signed); err == nil {
					t.Fatal("expired token accepted")
				}
			})
		}
	}
}

func TestIssuerVerificationRetainsRotationKeys(t *testing.T) {
	newIssuer := func(kid string, oldKeys map[string]string) *token.Issuer {
		t.Helper()
		_, private, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		issuer, err := token.NewIssuer(token.Config{Issuer: "identity", ActiveKeyID: kid, PrivateKeyB64: base64.RawURLEncoding.EncodeToString(private), VerifyKeys: oldKeys, TTL: time.Minute})
		if err != nil {
			t.Fatal(err)
		}
		return issuer
	}
	old := newIssuer("old", nil)
	rotated := newIssuer("new", map[string]string{"old": old.PublicKeys()[0].PublicKeyB64})
	for _, issuer := range []*token.Issuer{old, rotated} {
		signed, _, err := issuer.Issue(token.IssueRequest{Subject: "user", IdentityType: auth.IdentityHuman, Scope: auth.ScopeUser, AuthMethod: "test"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := rotated.Verify(signed); err != nil {
			t.Fatalf("rotation rejected token: %v", err)
		}
	}
}
