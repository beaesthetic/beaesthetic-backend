package server

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
)

func TestRoutePermissions(t *testing.T) {
	key := paseto.NewV4AsymmetricSecretKey()
	now := time.Now().UTC()
	verifier, err := auth.NewVerifier(auth.Config{Issuer: "identity", PublicKeys: map[string]string{"test": base64.RawURLEncoding.EncodeToString(key.Public().ExportBytes())}, Clock: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	router, err := New(&HttpHandlers{Customer: &Server{}, Fidelity: &Server{}, Wallet: &Server{}}, nil, verifier, "org-1")
	if err != nil {
		t.Fatal(err)
	}
	reached := false
	token := func(permission, org, scope string, expired bool) string {
		p := paseto.NewToken()
		p.SetIssuer("identity")
		p.SetSubject("person")
		p.SetIssuedAt(now.Add(-time.Minute))
		p.SetExpiration(now.Add(time.Hour))
		p.SetJti("jti")
		p.Set("kid", "test")
		p.Set("identity_type", "human")
		p.Set("auth_method", "oidc")
		p.Set("token_scope", scope)
		p.Set("organization_id", org)
		p.Set("membership_id", "member")
		p.Set("roles", []string{"owner"})
		p.Set("permissions", []string{permission})
		if scope == "user" {
			p.Set("organization_id", "")
			p.Set("membership_id", "")
			p.Set("roles", []string{})
		}
		if expired {
			p.SetExpiration(now)
		}
		return p.V4Sign(key, nil)
	}
	// Exercise the actual registered routes. A terminal middleware verifies context
	// and aborts before each generated endpoint is invoked.
	terminal := func(c *gin.Context) {
		claims, ok := auth.FromContext(c.Request.Context())
		if !ok || claims.OrganizationID != "org-1" {
			t.Error("missing verified claims")
		}
		reached = true
		c.Status(204)
		c.Abort()
	}
	secured := newRouter(&HttpHandlers{Customer: &Server{}, Fidelity: &Server{}, Wallet: &Server{}}, nil, authorize(verifier, "org-1"), terminal)
	for _, route := range secured.Routes() {
		if route.Path == "/health" {
			continue
		}
		path := strings.NewReplacer(":customerId", "1", ":cardId", "1", ":walletId", "1", ":voucherId", "1", ":direction", "next").Replace(route.Path)
		wantPermission := "customer.write"
		if route.Method == "GET" || route.Path == "/customers/search:byPhone" {
			wantPermission = "customer.read"
		}
		other := "customer.read"
		if wantPermission == other {
			other = "customer.write"
		}
		cases := []struct {
			name, bearer string
			status       int
		}{
			{"missing", "", 401}, {"invalid", "bad", 401}, {"expired", token(wantPermission, "org-1", "membership", true), 401},
			{"other permission", token(other, "org-1", "membership", false), 403}, {"wildcard", token("*", "org-1", "membership", false), 403},
			{"tenant", token(wantPermission, "org-2", "membership", false), 403}, {"user scope", token(wantPermission, "", "user", false), 403},
			{"allowed", token(wantPermission, "org-1", "membership", false), 204},
		}
		for _, tc := range cases {
			t.Run(route.Method+" "+route.Path+"/"+tc.name, func(t *testing.T) {
				reached = false
				r := httptest.NewRequest(route.Method, path, nil)
				if tc.bearer != "" {
					r.Header.Set("Authorization", "Bearer "+tc.bearer)
				}
				w := httptest.NewRecorder()
				secured.ServeHTTP(w, r)
				if tc.status != 204 {
					production := httptest.NewRecorder()
					router.ServeHTTP(production, r.Clone(r.Context()))
					if production.Code != tc.status {
						t.Fatalf("production status=%d want=%d", production.Code, tc.status)
					}
				}
				if w.Code != tc.status || reached != (tc.status == 204) {
					t.Fatalf("status=%d reached=%v want=%d", w.Code, reached, tc.status)
				}
				if tc.status == 401 && w.Header().Get("WWW-Authenticate") != "Bearer" {
					t.Error("missing Bearer challenge")
				}
			})
		}
	}
	w := httptest.NewRecorder()
	secured.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != 503 {
		t.Fatalf("health status=%d", w.Code)
	}
	if _, err := New(&HttpHandlers{}, nil, nil, "org-1"); err == nil {
		t.Error("nil verifier accepted")
	}
	if _, err := New(&HttpHandlers{}, nil, verifier, ""); err == nil {
		t.Error("empty organization accepted")
	}
}
