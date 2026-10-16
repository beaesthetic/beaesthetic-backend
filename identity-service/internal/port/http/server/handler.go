package server

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	identity "github.com/petretiandrea/beaesthetic-backend/core-contracts/identity"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/membership"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/token"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Config struct {
	InternalAPIKey string
}

func New(cfg Config, issuer *token.Issuer, exchange *application.ExchangeService, memberships membership.Repository) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.RedirectTrailingSlash = false
	router.Use(gin.Recovery())
	router.GET("/health", health)
	router.GET("/v1/internal/keys", func(ctx *gin.Context) {
		if !authorize(ctx.Request, cfg.InternalAPIKey) {
			unauthorized(ctx)
			return
		}
		keys := make([]*identity.VerificationKey, 0)
		for _, key := range issuer.PublicKeys() {
			keys = append(keys, &identity.VerificationKey{Kid: key.KeyID, PublicKeyB64: key.PublicKeyB64})
		}
		writeProtoJSON(ctx, http.StatusOK, &identity.GetVerificationKeysResponse{Keys: keys})
	})
	router.POST("/v1/internal/tokens", func(ctx *gin.Context) {
		if !authorize(ctx.Request, cfg.InternalAPIKey) {
			unauthorized(ctx)
			return
		}
		defer ctx.Request.Body.Close()
		var request identity.IssueInternalTokenRequest
		if err := decodeProtoJSON(ctx, &request); err != nil {
			badRequest(ctx, err)
			return
		}
		identityType, err := tokenIdentityType(request.GetIdentityType())
		if err != nil {
			badRequest(ctx, err)
			return
		}
		scope := token.ScopeUser
		if request.GetOrganizationId() != "" || request.GetMembershipId() != "" {
			scope = token.ScopeMembership
		}
		accessToken, claims, err := issuer.Issue(token.IssueRequest{Subject: request.GetSubject(), IdentityType: identityType, Scope: scope, OrganizationID: request.GetOrganizationId(), MembershipID: request.GetMembershipId(), Roles: request.GetRoles(), Permissions: request.GetPermissions(), AuthMethod: request.GetAuthMethod()})
		if err != nil {
			badRequest(ctx, err)
			return
		}
		writeProtoJSON(ctx, http.StatusCreated, &identity.IssueInternalTokenResponse{Token: accessTokenProto(accessToken, claims)})
	})
	router.POST("/v1/internal/tokens/verify", func(ctx *gin.Context) {
		if !authorize(ctx.Request, cfg.InternalAPIKey) {
			unauthorized(ctx)
			return
		}
		defer ctx.Request.Body.Close()
		var request identity.VerifyInternalTokenRequest
		if err := decodeProtoJSON(ctx, &request); err != nil {
			badRequest(ctx, err)
			return
		}
		claims, err := issuer.Verify(request.GetToken())
		if err != nil {
			writeJSON(ctx, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
			return
		}
		writeProtoJSON(ctx, http.StatusOK, &identity.VerifyInternalTokenResponse{Authorization: authorizationFromClaims(claims)})
	})
	router.POST("/oauth/token", func(ctx *gin.Context) {
		if exchange == nil {
			writeJSON(ctx, http.StatusServiceUnavailable, map[string]string{"error": "temporarily_unavailable"})
			return
		}
		defer ctx.Request.Body.Close()
		var request identity.ExchangeTokenRequest
		if err := decodeProtoJSON(ctx, &request); err != nil {
			badRequest(ctx, err)
			return
		}
		assertion := request.GetSubjectAssertion()
		if assertion == nil {
			badRequest(ctx, errors.New("subjectAssertion is required"))
			return
		}
		accessToken, claims, err := exchange.Exchange(ctx.Request.Context(), application.Request{
			GrantType: request.GetGrantType(), SubjectTokenType: assertion.GetTokenType(), Authenticator: assertion.GetAuthenticator(), Authorizer: request.GetAuthorizer(), SubjectToken: assertion.GetToken(),
			OrganizationID: request.GetOrganizationId(),
		})
		if err != nil {
			writeJSON(ctx, http.StatusUnauthorized, map[string]string{"error": "invalid_grant"})
			return
		}
		writeProtoJSON(ctx, http.StatusOK, &identity.ExchangeTokenResponse{Token: accessTokenProto(accessToken, claims)})
	})
	router.GET("/v1/me/organizations", func(ctx *gin.Context) {
		claims, ok := bearerClaims(ctx, issuer)
		if !ok {
			unauthorized(ctx)
			return
		}
		repository, ok := memberships.(membership.OrganizationRepository)
		if !ok {
			writeJSON(ctx, http.StatusServiceUnavailable, map[string]string{"error": "temporarily_unavailable"})
			return
		}
		organizations, err := repository.ListOrganizations(ctx.Request.Context(), claims.Subject)
		if err != nil {
			writeJSON(ctx, http.StatusInternalServerError, map[string]string{"error": "list_organizations"})
			return
		}
		response := make([]*identity.OrganizationMembership, 0, len(organizations))
		for _, organization := range organizations {
			response = append(response, &identity.OrganizationMembership{OrganizationId: organization.ID, OrganizationName: organization.Name, MembershipId: organization.MembershipID, Roles: organization.Roles})
		}
		writeProtoJSON(ctx, http.StatusOK, &identity.ListMyOrganizationsResponse{Organizations: response})
	})
	router.GET("/v1/organizations/:organization_id/authorization", func(ctx *gin.Context) {
		claims, ok := bearerClaims(ctx, issuer)
		if !ok {
			unauthorized(ctx)
			return
		}
		organizationID := ctx.Param("organization_id")
		record, err := memberships.FindActiveMembership(ctx.Request.Context(), claims.Subject, organizationID)
		if err != nil || !record.Active {
			writeJSON(ctx, http.StatusNotFound, map[string]string{"error": "authorization_not_found"})
			return
		}
		writeProtoJSON(ctx, http.StatusOK, &identity.GetMyAuthorizationResponse{Authorization: membershipAuthorization(record)})
	})
	router.GET("/v1/organizations/:organization_id/members", func(ctx *gin.Context) {
		management, claims, ok := managementAccess(ctx, issuer, memberships, "identity.memberships.read")
		if !ok {
			return
		}
		members, err := management.ListMembers(ctx.Request.Context(), ctx.Param("organization_id"))
		if err != nil {
			writeJSON(ctx, http.StatusInternalServerError, map[string]string{"error": "list_members"})
			return
		}
		_ = claims
		response := make([]*identity.OrganizationMember, 0, len(members))
		for _, member := range members {
			response = append(response, memberProto(member))
		}
		writeProtoJSON(ctx, http.StatusOK, &identity.ListOrganizationMembersResponse{Members: response})
	})
	router.POST("/v1/organizations/:organization_id/members", func(ctx *gin.Context) {
		management, claims, ok := managementAccess(ctx, issuer, memberships, "identity.memberships.write")
		if !ok {
			return
		}
		var request identity.AddOrganizationMemberRequest
		if err := decodeProtoJSON(ctx, &request); err != nil {
			badRequest(ctx, err)
			return
		}
		if containsRole(request.GetRoles(), "owner") && !containsRole(claims.Roles, "owner") {
			forbidden(ctx)
			return
		}
		member, err := management.AddMember(ctx.Request.Context(), ctx.Param("organization_id"), request.GetEmail(), request.GetRoles())
		if err != nil {
			badRequest(ctx, err)
			return
		}
		writeProtoJSON(ctx, http.StatusCreated, &identity.AddOrganizationMemberResponse{Member: memberProto(member)})
	})
	router.PUT("/v1/organizations/:organization_id/members/:user_id/roles", func(ctx *gin.Context) {
		management, claims, ok := managementAccess(ctx, issuer, memberships, "identity.memberships.write")
		if !ok {
			return
		}
		var request identity.ReplaceMemberRolesRequest
		if err := decodeProtoJSON(ctx, &request); err != nil {
			badRequest(ctx, err)
			return
		}
		if containsRole(request.GetRoles(), "owner") && !containsRole(claims.Roles, "owner") {
			forbidden(ctx)
			return
		}
		member, err := management.ReplaceMemberRoles(ctx.Request.Context(), ctx.Param("organization_id"), ctx.Param("user_id"), request.GetRoles())
		if err != nil {
			badRequest(ctx, err)
			return
		}
		writeProtoJSON(ctx, http.StatusOK, &identity.ReplaceMemberRolesResponse{Member: memberProto(member)})
	})
	router.DELETE("/v1/organizations/:organization_id/members/:user_id", func(ctx *gin.Context) {
		management, claims, ok := managementAccess(ctx, issuer, memberships, "identity.memberships.write")
		if !ok {
			return
		}
		target, err := memberships.FindActiveMembership(ctx.Request.Context(), ctx.Param("user_id"), ctx.Param("organization_id"))
		if err != nil || !target.Active {
			writeJSON(ctx, http.StatusNotFound, map[string]string{"error": "membership_not_found"})
			return
		}
		if containsRole(target.Roles, "owner") && !containsRole(claims.Roles, "owner") {
			forbidden(ctx)
			return
		}
		if err := management.RemoveMember(ctx.Request.Context(), ctx.Param("organization_id"), ctx.Param("user_id")); err != nil {
			writeJSON(ctx, http.StatusNotFound, map[string]string{"error": "membership_not_found"})
			return
		}
		ctx.Status(http.StatusNoContent)
	})
	router.GET("/v1/roles", func(ctx *gin.Context) {
		if _, ok := bearerClaims(ctx, issuer); !ok {
			unauthorized(ctx)
			return
		}
		management, ok := memberships.(membership.ManagementRepository)
		if !ok {
			writeJSON(ctx, http.StatusServiceUnavailable, map[string]string{"error": "temporarily_unavailable"})
			return
		}
		roles, err := management.ListGlobalRoles(ctx.Request.Context())
		if err != nil {
			writeJSON(ctx, http.StatusInternalServerError, map[string]string{"error": "list_roles"})
			return
		}
		response := make([]*identity.Role, 0, len(roles))
		for _, role := range roles {
			response = append(response, &identity.Role{Name: role})
		}
		writeProtoJSON(ctx, http.StatusOK, &identity.ListRolesResponse{Roles: response})
	})
	router.GET("/v1/internal/organizations/:organization_id/users/:user_id/authorization", func(ctx *gin.Context) {
		if !authorize(ctx.Request, cfg.InternalAPIKey) {
			unauthorized(ctx)
			return
		}
		if memberships == nil {
			writeJSON(ctx, http.StatusServiceUnavailable, map[string]string{"error": "temporarily_unavailable"})
			return
		}
		userID, organizationID := ctx.Param("user_id"), ctx.Param("organization_id")
		record, err := memberships.FindActiveMembership(ctx.Request.Context(), userID, organizationID)
		if err != nil || !record.Active {
			writeJSON(ctx, http.StatusNotFound, map[string]string{"error": "authorization_not_found"})
			return
		}
		writeProtoJSON(ctx, http.StatusOK, &identity.GetEffectiveAuthorizationResponse{Authorization: membershipAuthorization(record)})
	})
	return router
}

func decodeProtoJSON(ctx *gin.Context, message proto.Message) error {
	body, err := io.ReadAll(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 32<<10))
	if err != nil {
		return err
	}
	return protojson.UnmarshalOptions{DiscardUnknown: false}.Unmarshal(body, message)
}

func writeProtoJSON(ctx *gin.Context, statusCode int, message proto.Message) {
	body, err := protojson.MarshalOptions{UseProtoNames: false, EmitUnpopulated: false}.Marshal(message)
	if err != nil {
		writeJSON(ctx, http.StatusInternalServerError, map[string]string{"error": "encode response"})
		return
	}
	ctx.Data(statusCode, "application/json", body)
}

func accessTokenProto(value string, claims token.Claims) *identity.AccessToken {
	return &identity.AccessToken{AccessToken: value, TokenType: "Bearer", ExpiresIn: int64(claims.ExpiresAt.Sub(claims.IssuedAt).Seconds()), ExpiresAt: timestamppb.New(claims.ExpiresAt)}
}

func authorizationFromClaims(claims token.Claims) *identity.EffectiveAuthorization {
	return &identity.EffectiveAuthorization{UserId: claims.Subject, MembershipId: claims.MembershipID, OrganizationId: claims.OrganizationID, Roles: claims.Roles, Permissions: claims.Permissions, IdentityType: protobufIdentityType(claims.IdentityType)}
}

func membershipAuthorization(record membership.Membership) *identity.EffectiveAuthorization {
	return &identity.EffectiveAuthorization{UserId: record.UserID, MembershipId: record.ID, OrganizationId: record.OrganizationID, Roles: record.Roles, Permissions: record.Permissions, IdentityType: identity.IdentityType_IDENTITY_TYPE_HUMAN}
}

func tokenIdentityType(value identity.IdentityType) (token.IdentityType, error) {
	switch value {
	case identity.IdentityType_IDENTITY_TYPE_HUMAN:
		return token.IdentityHuman, nil
	case identity.IdentityType_IDENTITY_TYPE_SERVICE:
		return token.IdentityService, nil
	case identity.IdentityType_IDENTITY_TYPE_INTEGRATION:
		return token.IdentityIntegration, nil
	default:
		return "", fmt.Errorf("invalid identityType")
	}
}

func protobufIdentityType(value token.IdentityType) identity.IdentityType {
	switch value {
	case token.IdentityService:
		return identity.IdentityType_IDENTITY_TYPE_SERVICE
	case token.IdentityIntegration:
		return identity.IdentityType_IDENTITY_TYPE_INTEGRATION
	default:
		return identity.IdentityType_IDENTITY_TYPE_HUMAN
	}
}

func health(ctx *gin.Context) {
	writeJSON(ctx, http.StatusOK, map[string]string{"status": "ok"})
}

func authorize(r *http.Request, expected string) bool {
	provided := r.Header.Get("X-Internal-API-Key")
	return expected != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func unauthorized(ctx *gin.Context) {
	writeJSON(ctx, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

func forbidden(ctx *gin.Context) {
	writeJSON(ctx, http.StatusForbidden, map[string]string{"error": "forbidden"})
}

func bearerClaims(ctx *gin.Context, issuer *token.Issuer) (token.Claims, bool) {
	value := ctx.GetHeader("Authorization")
	if len(value) <= len("Bearer ") || value[:len("Bearer ")] != "Bearer " {
		return token.Claims{}, false
	}
	claims, err := issuer.Verify(value[len("Bearer "):])
	return claims, err == nil
}

func hasPermission(claims token.Claims, permission string) bool {
	for _, value := range claims.Permissions {
		if value == permission {
			return true
		}
	}
	return false
}

func containsRole(roles []string, expected string) bool {
	for _, role := range roles {
		if role == expected {
			return true
		}
	}
	return false
}

func managementAccess(ctx *gin.Context, issuer *token.Issuer, repository membership.Repository, permission string) (membership.ManagementRepository, token.Claims, bool) {
	claims, ok := bearerClaims(ctx, issuer)
	management, supported := repository.(membership.ManagementRepository)
	if !ok || !supported || claims.Scope != token.ScopeMembership || !hasPermission(claims, permission) || claims.OrganizationID != ctx.Param("organization_id") {
		unauthorized(ctx)
		return nil, token.Claims{}, false
	}
	return management, claims, true
}

func memberProto(member membership.Member) *identity.OrganizationMember {
	return &identity.OrganizationMember{UserId: member.UserID, Email: member.Email, MembershipId: member.MembershipID, Active: member.Active, Roles: member.Roles}
}

func badRequest(ctx *gin.Context, err error) {
	message := "invalid request"
	if errors.Is(err, http.ErrBodyReadAfterClose) {
		message = "invalid request body"
	}
	writeJSON(ctx, http.StatusBadRequest, map[string]string{"error": message})
}

func writeJSON(ctx *gin.Context, status int, value any) {
	ctx.JSON(status, value)
}
