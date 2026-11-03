package auth

import (
	"context"
	"net/http"
	"strings"
)

type claimsKey struct{}

// FromContext returns a copy so downstream code cannot mutate authentication state.
func FromContext(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(claimsKey{}).(Claims)
	c.Roles = append([]string(nil), c.Roles...)
	c.Permissions = append([]string(nil), c.Permissions...)
	return c, ok
}

// Authenticate requires exactly one Authorization: Bearer <token> header.
func (v *Verifier) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := r.Header.Values("Authorization")
		if len(headers) != 1 {
			unauthorized(w)
			return
		}
		fields := strings.Fields(headers[0])
		if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
			unauthorized(w)
			return
		}
		claims, err := v.Verify(fields[1])
		if err != nil {
			unauthorized(w)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claimsKey{}, claims)))
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// Policy requires all permissions. Optional scope and organization constraints are
// checked against verified claims. OrganizationID must come from trusted routing,
// never from the token itself. An empty policy requires authentication only.
type Policy struct {
	Permissions    []string
	Scope          Scope
	OrganizationID string
}

func (p Policy) Allows(c Claims) bool {
	if p.Scope != "" && (!p.Scope.Valid() || c.Scope != p.Scope) {
		return false
	}
	if p.OrganizationID != "" && (c.Scope != ScopeMembership || c.OrganizationID != p.OrganizationID) {
		return false
	}
	for _, permission := range p.Permissions {
		if !c.HasPermission(permission) {
			return false
		}
	}
	return true
}

// Require must run after Authenticate. Missing authentication is 401; a failed
// policy is 403. The policy is copied to prevent later caller mutations.
func Require(policy Policy) func(http.Handler) http.Handler {
	policy.Permissions = append([]string(nil), policy.Permissions...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, ok := FromContext(r.Context())
			if !ok {
				unauthorized(w)
				return
			}
			if !policy.Allows(c) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
