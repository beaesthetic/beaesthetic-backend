package server

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
)

const organizationHeader = "X-Organization-ID"

// The requested organization must match the verified membership token.
func authorize(verifier *auth.Verifier, permissions ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accepted := false
		handler := verifier.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			values := r.Header.Values(organizationHeader)
			if len(values) != 1 || values[0] == "" || strings.ContainsAny(values[0], " ,\t\r\n") {
				http.Error(w, "Exactly one X-Organization-ID header is required", http.StatusBadRequest)
				return
			}
			policy := auth.Policy{Permissions: permissions, Scope: auth.ScopeMembership, OrganizationID: values[0]}
			auth.Require(policy)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx.Request = r
				accepted = true
			})).ServeHTTP(w, r)
		}))
		handler.ServeHTTP(ctx.Writer, ctx.Request)
		if !accepted {
			ctx.Abort()
		}
	}
}
