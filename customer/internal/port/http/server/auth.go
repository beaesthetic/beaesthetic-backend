package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
)

// All customer, fidelity and wallet routes share the customer permissions.
// The database belongs to the configured organization; claims never select it.
func authorize(verifier *auth.Verifier, organizationID string, permissions ...string) gin.HandlerFunc {
	policy := auth.Policy{Permissions: permissions, Scope: auth.ScopeMembership, OrganizationID: organizationID}
	return func(ctx *gin.Context) {
		accepted := false
		handler := verifier.Authenticate(auth.Require(policy)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx.Request = r
			accepted = true
		})))
		handler.ServeHTTP(ctx.Writer, ctx.Request)
		if !accepted {
			ctx.Abort()
			return
		}
	}
}
