package server

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
	customerapi "github.com/petretiandrea/beaesthetic-backend/customer/internal/port/http/server/customer"
)

func TestRoutePermissions(t *testing.T) {
	key := paseto.NewV4AsymmetricSecretKey()
	now := time.Now().UTC()
	verifier, err := auth.NewVerifier(auth.Config{Issuer: "identity", PublicKeys: map[string]string{"test": base64.RawURLEncoding.EncodeToString(key.Public().ExportBytes())}, Clock: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	readHandler := &customerReadStub{}
	router, err := New(&HttpHandlers{Customer: readHandler, Fidelity: &Server{}, Wallet: &Server{}}, nil, verifier)
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
		if !ok || claims.OrganizationID == "" || claims.Scope != auth.ScopeMembership {
			t.Error("missing verified claims")
		}
		reached = true
		c.Status(204)
		c.Abort()
	}
	secured := newRouter(&HttpHandlers{Customer: &Server{}, Fidelity: &Server{}, Wallet: &Server{}}, nil, func(permissions ...string) gin.HandlerFunc {
		middleware := authorize(verifier, permissions...)
		return func(c *gin.Context) {
			middleware(c)
			if !c.IsAborted() {
				terminal(c)
			}
		}
	})
	for _, route := range secured.Routes() {
		if route.Path == "/health" || route.Path == "/internal/customers/:customerId" {
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
			headers      []string
		}{

			{"missing", "", 401, nil},
			{"invalid", "bad", 401, nil},
			{"expired", token(wantPermission, "org-1", "membership", true), 401, nil},
			{"other permission", token(other, "org-1", "membership", false), 403, nil},
			{"wildcard", token("*", "org-1", "membership", false), 403, nil},
			{"user scope", token(wantPermission, "", "user", false), 403, nil},
			{"membership without organization", token(wantPermission, "", "membership", false), 401, nil},
			{"allowed without organization header", token(wantPermission, "org-1", "membership", false), 204, nil},
			{"another token organization", token(wantPermission, "org-2", "membership", false), 204, nil},
			{"organization header ignored", token(wantPermission, "org-1", "membership", false), 204, []string{"org-2"}},
			{"empty organization header ignored", token(wantPermission, "org-1", "membership", false), 204, []string{""}},
			{"duplicate organization header ignored", token(wantPermission, "org-1", "membership", false), 204, []string{"org-2", "org-3"}},
		}
		for _, tc := range cases {
			t.Run(route.Method+" "+route.Path+"/"+tc.name, func(t *testing.T) {
				reached = false
				r := httptest.NewRequest(route.Method, path, nil)
				if tc.bearer != "" {
					r.Header.Set("Authorization", "Bearer "+tc.bearer)
				}
				for _, value := range tc.headers {
					r.Header.Add("X-Organization-ID", value)
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
	// Exercise generated binding and the strict handler without an organization header.
	for _, headers := range [][]string{nil, {"unrelated-org"}, {"org-2", "org-3"}} {
		readHandler.called = false
		request := httptest.NewRequest(http.MethodGet, "/admin/customers", nil)
		request.Header.Set("Authorization", "Bearer "+token("customer.read", "org-1", "membership", false))
		for _, value := range headers {
			request.Header.Add("X-Organization-ID", value)
		}
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != 200 || !readHandler.called || readHandler.claims.OrganizationID != "org-1" {
			t.Fatalf("status=%d called=%v claims=%+v", response.Code, readHandler.called, readHandler.claims)
		}
	}
	w := httptest.NewRecorder()
	secured.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != 503 {
		t.Fatalf("health status=%d", w.Code)
	}
	if _, err := New(&HttpHandlers{}, nil, nil); err == nil {
		t.Error("nil verifier accepted")
	}
}

// Embedding supplies unused operations; only the read path is invoked here.
type customerReadStub struct {
	*Server
	called bool
	claims auth.Claims
}

func (s *customerReadStub) GetAllCustomers(ctx context.Context, _ customerapi.GetAllCustomersRequestObject) (customerapi.GetAllCustomersResponseObject, error) {
	s.called = true
	s.claims, _ = auth.FromContext(ctx)
	return customerapi.GetAllCustomers200JSONResponse{}, nil
}

func (s *customerReadStub) GetCustomerById(_ context.Context, _ customerapi.GetCustomerByIdRequestObject) (customerapi.GetCustomerByIdResponseObject, error) {
	return customerapi.GetCustomerById200JSONResponse{Id: "customer-1", Name: "Ada", Surname: "Lovelace"}, nil
}

func (s *customerReadStub) GetInternalCustomerById(_ context.Context, _ customerapi.GetInternalCustomerByIdRequestObject) (customerapi.GetInternalCustomerByIdResponseObject, error) {
	return customerapi.GetInternalCustomerById200JSONResponse{Id: "customer-1", Name: "Ada", Surname: "Lovelace"}, nil
}
