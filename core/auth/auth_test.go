package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
)

var testNow = time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)

func fixture(t *testing.T) (*Verifier, paseto.V4AsymmetricSecretKey) {
	t.Helper()
	key := paseto.NewV4AsymmetricSecretKey()
	v, err := NewVerifier(Config{Issuer: "identity", PublicKeys: map[string]string{"active": base64.RawURLEncoding.EncodeToString(key.Public().ExportBytes())}, Clock: func() time.Time { return testNow }})
	if err != nil {
		t.Fatal(err)
	}
	return v, key
}

// Mirror the claim setters in identity-service/internal/domain/token/issuer.go.
func signedToken(key paseto.V4AsymmetricSecretKey, edit func(*paseto.Token)) string {
	token := paseto.NewToken()
	token.SetIssuer("identity")
	token.SetSubject("person-1")
	token.SetIssuedAt(testNow.Add(-time.Minute))
	token.SetExpiration(testNow.Add(time.Minute))
	token.SetJti("token-1")
	token.Set("kid", "active")
	token.Set("identity_type", "human")
	token.Set("token_scope", "membership")
	token.Set("organization_id", "org-1")
	token.Set("membership_id", "member-1")
	token.Set("roles", []string{"admin"})
	token.Set("permissions", []string{"appointments.read"})
	token.Set("auth_method", "provider:google")
	if edit != nil {
		edit(&token)
	}
	return token.V4Sign(key, nil)
}

func TestVerifier(t *testing.T) {
	v, key := fixture(t)
	valid := signedToken(key, nil)
	claims, err := v.Verify(valid)
	if err != nil || claims.Subject != "person-1" || claims.OrganizationID != "org-1" || !claims.HasPermission("appointments.read") {
		t.Fatalf("claims=%+v err=%v", claims, err)
	}
	cases := map[string]func(*paseto.Token){
		"issuer":               func(p *paseto.Token) { p.SetIssuer("other") },
		"kid":                  func(p *paseto.Token) { p.Set("kid", "unknown") },
		"expired":              func(p *paseto.Token) { p.SetExpiration(testNow) },
		"future issued":        func(p *paseto.Token) { p.SetIssuedAt(testNow.Add(time.Second)) },
		"future nbf":           func(p *paseto.Token) { p.SetNotBefore(testNow.Add(time.Second)) },
		"empty subject":        func(p *paseto.Token) { p.SetSubject("") },
		"empty jti":            func(p *paseto.Token) { p.SetJti("") },
		"identity type":        func(p *paseto.Token) { p.Set("identity_type", "unknown") },
		"scope":                func(p *paseto.Token) { p.Set("token_scope", "unknown") },
		"membership":           func(p *paseto.Token) { p.Set("membership_id", "") },
		"organization":         func(p *paseto.Token) { p.Set("organization_id", "") },
		"user with membership": func(p *paseto.Token) { p.Set("token_scope", "user") },
		"permissions type":     func(p *paseto.Token) { p.Set("permissions", "admin") },
		"auth method":          func(p *paseto.Token) { p.Set("auth_method", "") },
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := v.Verify(signedToken(key, edit)); err == nil {
				t.Fatal("invalid token accepted")
			}
		})
	}
	for _, value := range []string{"", "v4.local.invalid", valid + "x", signedToken(paseto.NewV4AsymmetricSecretKey(), nil)} {
		if _, err := v.Verify(value); err == nil {
			t.Fatal("invalid signature accepted")
		}
	}
	user := signedToken(key, func(p *paseto.Token) {
		p.Set("token_scope", "user")
		p.Set("organization_id", "")
		p.Set("membership_id", "")
		p.Set("roles", []string{})
		p.Set("identity_type", "service")
	})
	if _, err := v.Verify(user); err != nil {
		t.Fatal(err)
	}
}

func TestRotationAndConfig(t *testing.T) {
	v, key := fixture(t)
	second := paseto.NewV4AsymmetricSecretKey()
	rotated, err := NewVerifier(Config{Issuer: "identity", PublicKeys: map[string]string{"active": base64.RawURLEncoding.EncodeToString(key.Public().ExportBytes()), "old": base64.RawURLEncoding.EncodeToString(second.Public().ExportBytes())}, Clock: v.clock})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rotated.Verify(signedToken(second, func(p *paseto.Token) { p.Set("kid", "old") })); err != nil {
		t.Fatal(err)
	}
	for _, cfg := range []Config{{}, {Issuer: "identity"}, {Issuer: "identity", PublicKeys: map[string]string{"": "x"}}, {Issuer: "identity", PublicKeys: map[string]string{"key": "%%%"}}, {Issuer: "identity", PublicKeys: map[string]string{"key": "eA"}}} {
		if _, err := NewVerifier(cfg); err == nil {
			t.Fatal("invalid config accepted")
		}
	}
}

func TestServiceConfig(t *testing.T) {
	_, key := fixture(t)
	encoded := base64.RawURLEncoding.EncodeToString(key.Public().ExportBytes())
	if _, err := NewVerifier(Config{Issuer: "identity", PublicKeysJSON: `{"active":"` + encoded + `"}`}); err != nil {
		t.Fatal(err)
	}
	for _, cfg := range []Config{{}, {Issuer: "identity", PublicKeysJSON: "{}"}, {Issuer: "identity", PublicKeysJSON: "bad"}, {Issuer: "identity", PublicKeysJSON: `{"key":"bad"}`}} {
		if _, err := NewVerifier(cfg); err == nil {
			t.Errorf("accepted invalid service config: %+v", cfg)
		}
	}
}

func TestHTTPAuthorization(t *testing.T) {
	v, key := fixture(t)
	token := signedToken(key, nil)
	for _, tc := range []struct {
		name    string
		headers []string
		policy  Policy
		want    int
	}{
		{"missing", nil, Policy{}, 401},
		{"scheme", []string{"Basic " + token}, Policy{}, 401},
		{"empty bearer", []string{"Bearer"}, Policy{}, 401},
		{"duplicate", []string{"Bearer " + token, "Bearer " + token}, Policy{}, 401},
		{"tampered", []string{"Bearer " + token + "x"}, Policy{}, 401},
		{"allowed", []string{"bearer " + token}, Policy{Permissions: []string{"appointments.read"}, Scope: ScopeMembership, OrganizationID: "org-1"}, 204},
		{"missing permission", []string{"Bearer " + token}, Policy{Permissions: []string{"appointments.write"}}, 403},
		{"all permissions", []string{"Bearer " + token}, Policy{Permissions: []string{"appointments.read", "appointments.write"}}, 403},
		{"tenant", []string{"Bearer " + token}, Policy{OrganizationID: "org-2"}, 403},
		{"scope", []string{"Bearer " + token}, Policy{Scope: ScopeUser}, 403},
		{"bad scope", []string{"Bearer " + token}, Policy{Scope: "invalid"}, 403},
		{"empty permission", []string{"Bearer " + token}, Policy{Permissions: []string{""}}, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			handler := v.Authenticate(Require(tc.policy)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				c, ok := FromContext(r.Context())
				if !ok || c.Subject != "person-1" {
					t.Fatal("missing claims")
				}
				c.Permissions[0] = "mutated"
				again, _ := FromContext(r.Context())
				if again.Permissions[0] != "appointments.read" {
					t.Fatal("context mutated")
				}
				w.WriteHeader(204)
			})))
			req := httptest.NewRequest("GET", "/", nil)
			for _, header := range tc.headers {
				req.Header.Add("Authorization", header)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != tc.want || called != (tc.want == 204) {
				t.Fatalf("status=%d called=%v", w.Code, called)
			}
			if tc.want == 401 && w.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatal("missing challenge")
			}
		})
	}
	w := httptest.NewRecorder()
	Require(Policy{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("unauthenticated handler called") })).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
	if (Claims{Roles: []string{"admin"}, Permissions: []string{"*"}}).HasPermission("appointments.write") {
		t.Fatal("implicit permission grant")
	}
}

func TestIdentityIssuerCompatibility(t *testing.T) {
	data, err := os.ReadFile("testdata/identity-token.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		PublicKey string `json:"public_key"`
		Token     string `json:"token"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	v, err := NewVerifier(Config{Issuer: "identity", PublicKeys: map[string]string{"fixture": fixture.PublicKey}, Clock: func() time.Time { return testNow.Add(time.Minute) }})
	if err != nil {
		t.Fatal(err)
	}
	c, err := v.Verify(fixture.Token)
	if err != nil {
		t.Fatal(err)
	}
	if c.Subject != "person-1" || c.OrganizationID != "org-1" || c.MembershipID != "member-1" || c.Scope != ScopeMembership || !c.HasPermission("appointments.read") {
		t.Fatalf("unexpected Identity claims: %+v", c)
	}
}

func TestPublicKeyAliases(t *testing.T) {
	_, key := fixture(t)
	encoded := base64.RawURLEncoding.EncodeToString(key.Public().ExportBytes())
	v, err := NewVerifier(Config{Issuer: "identity", PublicKeys: map[string]string{"active": encoded, "alias": encoded}, Clock: func() time.Time { return testNow }})
	if err != nil {
		t.Fatal(err)
	}
	for range 50 {
		if _, err := v.Verify(signedToken(key, nil)); err != nil {
			t.Fatal(err)
		}
	}
}
