package server

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
	"go.uber.org/zap"
)

func TestCalendarRoutesRequirePermissions(t *testing.T) {
	key := paseto.NewV4AsymmetricSecretKey()
	now := time.Date(2026, time.November, 14, 0, 0, 0, 0, time.UTC)
	verifier, err := auth.NewVerifier(auth.Config{
		Issuer: "identity",
		PublicKeys: map[string]string{"test": base64.RawURLEncoding.EncodeToString(key.Public().ExportBytes())},
		Clock:      func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	token := func(permission string) string {
		p := paseto.NewToken()
		p.SetIssuer("identity")
		p.SetSubject("user-1")
		p.SetJti("token-1")
		p.SetIssuedAt(now.Add(-time.Minute))
		p.SetExpiration(now.Add(time.Hour))
		p.Set("kid", "test")
		p.Set("identity_type", "user")
		p.Set("token_scope", "membership")
		p.Set("membership_id", "member-1")
		p.Set("organization_id", "org-1")
		p.Set("auth_method", "oidc")
		p.Set("permissions", []string{permission})
		return p.V4Sign(key, nil)
	}

	router, err := New(&HttpHandlers{
		Calendar: &Server{},
		HealthChecker: func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) },
	}, zap.NewNop(), verifier)
	if err != nil {
		t.Fatal(err)
	}

	for _, route := range router.Routes() {
		if route.Path == "/health" {
			request := httptest.NewRequest(http.MethodGet, route.Path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusNoContent {
				t.Errorf("health status = %d, want %d", response.Code, http.StatusNoContent)
			}
			continue
		}

		path := route.Path
		for _, replacement := range []struct{ old, new string }{{":calendar_event_id", "event-1"}, {":id", "event-1"}} {
			path = replacePathParameter(path, replacement.old, replacement.new)
		}
		request := httptest.NewRequest(route.Method, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without token = %d, want %d", route.Method, route.Path, response.Code, http.StatusUnauthorized)
		}

		wrongPermission := "appointment.write"
		if route.Method != http.MethodGet {
			wrongPermission = "appointment.read"
		}
		request = httptest.NewRequest(route.Method, path, nil)
		request.Header.Set("Authorization", "Bearer "+token(wrongPermission))
		response = httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Errorf("%s %s with %s = %d, want %d", route.Method, route.Path, wrongPermission, response.Code, http.StatusForbidden)
		}
	}
}

func replacePathParameter(path, old, new string) string {
	for i := 0; i+len(old) <= len(path); i++ {
		if path[i:i+len(old)] == old {
			return path[:i] + new + path[i+len(old):]
		}
	}
	return path
}
