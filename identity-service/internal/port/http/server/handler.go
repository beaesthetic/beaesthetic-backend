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
		accessToken, claims, err := issuer.Issue(token.IssueRequest{Subject: request.GetSubject(), IdentityType: identityType, Audience: request.GetAudience(), OrganizationID: request.GetOrganizationId(), MembershipID: request.GetMembershipId(), Roles: request.GetRoles(), Permissions: request.GetPermissions(), AuthMethod: request.GetAuthMethod()})
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
		claims, err := issuer.Verify(request.GetToken(), request.GetAudience())
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
			Audience: request.GetAudience(), OrganizationID: request.GetOrganizationId(),
		})
		if err != nil {
			writeJSON(ctx, http.StatusUnauthorized, map[string]string{"error": "invalid_grant"})
			return
		}
		writeProtoJSON(ctx, http.StatusOK, &identity.ExchangeTokenResponse{Token: accessTokenProto(accessToken, claims)})
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
