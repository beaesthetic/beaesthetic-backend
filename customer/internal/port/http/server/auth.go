package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
)

// Customer currently serves a single organization without tenant-partitioned data.
// Organization claims remain available in context, but no organization header or
// organization constraint is applied here.
func authorize(verifier *auth.Verifier, permissions ...string) gin.HandlerFunc {
	policy := auth.Policy{Permissions: permissions, Scope: auth.ScopeMembership}
	require := auth.Require(policy)
	return func(ctx *gin.Context) {
		accepted := false
		handler := verifier.Authenticate(require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx.Request = r
			accepted = true
		})))
		handler.ServeHTTP(ctx.Writer, ctx.Request)
		if !accepted {
			ctx.Abort()
		}
	}
}
