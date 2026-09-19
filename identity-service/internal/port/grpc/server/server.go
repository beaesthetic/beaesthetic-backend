package server

import (
	"context"
	"crypto/subtle"
	"fmt"

	identity "github.com/petretiandrea/beaesthetic-backend/core-contracts/identity"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/application"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/membership"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/token"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TokenServer struct {
	identity.UnimplementedTokenServiceServer
	internalAPIKey string
	issuer         *token.Issuer
	exchange       *application.ExchangeService
}

func NewTokenServer(internalAPIKey string, issuer *token.Issuer, exchange *application.ExchangeService) *TokenServer {
	return &TokenServer{internalAPIKey: internalAPIKey, issuer: issuer, exchange: exchange}
}

func (s *TokenServer) ExchangeToken(ctx context.Context, request *identity.ExchangeTokenRequest) (*identity.ExchangeTokenResponse, error) {
	if s.exchange == nil {
		return nil, status.Error(codes.Unavailable, "token exchange is unavailable")
	}
	assertion := request.GetSubjectAssertion()
	if assertion == nil {
		return nil, status.Error(codes.InvalidArgument, "subject_assertion is required")
	}
	accessToken, claims, err := s.exchange.Exchange(ctx, application.Request{
		GrantType: request.GetGrantType(), SubjectTokenType: assertion.GetTokenType(), Authenticator: assertion.GetAuthenticator(), Authorizer: request.GetAuthorizer(), SubjectToken: assertion.GetToken(),
		Audience: request.GetAudience(), OrganizationID: request.GetOrganizationId(),
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token exchange request")
	}
	return &identity.ExchangeTokenResponse{Token: accessTokenResponse(accessToken, claims)}, nil
}

func (s *TokenServer) GetVerificationKeys(ctx context.Context, _ *identity.GetVerificationKeysRequest) (*identity.GetVerificationKeysResponse, error) {
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}
	keys := s.issuer.PublicKeys()
	response := make([]*identity.VerificationKey, 0, len(keys))
	for _, key := range keys {
		response = append(response, &identity.VerificationKey{Kid: key.KeyID, PublicKeyB64: key.PublicKeyB64})
	}
	return &identity.GetVerificationKeysResponse{Keys: response}, nil
}

func (s *TokenServer) IssueInternalToken(ctx context.Context, request *identity.IssueInternalTokenRequest) (*identity.IssueInternalTokenResponse, error) {
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}
	identityType, err := identityType(request.GetIdentityType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	accessToken, claims, err := s.issuer.Issue(token.IssueRequest{Subject: request.GetSubject(), IdentityType: identityType, Audience: request.GetAudience(), OrganizationID: request.GetOrganizationId(), MembershipID: request.GetMembershipId(), Roles: request.GetRoles(), Permissions: request.GetPermissions(), AuthMethod: request.GetAuthMethod()})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token request")
	}
	return &identity.IssueInternalTokenResponse{Token: accessTokenResponse(accessToken, claims)}, nil
}

func (s *TokenServer) VerifyInternalToken(ctx context.Context, request *identity.VerifyInternalTokenRequest) (*identity.VerifyInternalTokenResponse, error) {
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}
	claims, err := s.issuer.Verify(request.GetToken(), request.GetAudience())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return &identity.VerifyInternalTokenResponse{Authorization: authorization(claims)}, nil
}

func (s *TokenServer) authorize(ctx context.Context) error {
	values := metadata.ValueFromIncomingContext(ctx, "x-internal-api-key")
	if len(values) != 1 || s.internalAPIKey == "" || subtle.ConstantTimeCompare([]byte(values[0]), []byte(s.internalAPIKey)) != 1 {
		return status.Error(codes.Unauthenticated, "unauthorized")
	}
	return nil
}

type MembershipServer struct {
	identity.UnimplementedMembershipServiceServer
	internalAPIKey string
	repository     membership.Repository
}

func NewMembershipServer(internalAPIKey string, repository membership.Repository) *MembershipServer {
	return &MembershipServer{internalAPIKey: internalAPIKey, repository: repository}
}

func (s *MembershipServer) GetEffectiveAuthorization(ctx context.Context, request *identity.GetEffectiveAuthorizationRequest) (*identity.GetEffectiveAuthorizationResponse, error) {
	if err := (&TokenServer{internalAPIKey: s.internalAPIKey}).authorize(ctx); err != nil {
		return nil, err
	}
	if request.GetUserId() == "" || request.GetOrganizationId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and organization_id are required")
	}
	record, err := s.repository.FindActiveMembership(ctx, request.GetUserId(), request.GetOrganizationId())
	if err != nil || !record.Active {
		return nil, status.Error(codes.NotFound, "authorization not found")
	}
	return &identity.GetEffectiveAuthorizationResponse{Authorization: membershipAuthorization(record)}, nil
}

func accessTokenResponse(value string, claims token.Claims) *identity.AccessToken {
	return &identity.AccessToken{AccessToken: value, TokenType: "Bearer", ExpiresIn: int64(claims.ExpiresAt.Sub(claims.IssuedAt).Seconds()), ExpiresAt: timestamppb.New(claims.ExpiresAt)}
}

func authorization(claims token.Claims) *identity.EffectiveAuthorization {
	return &identity.EffectiveAuthorization{UserId: claims.Subject, MembershipId: claims.MembershipID, OrganizationId: claims.OrganizationID, Roles: claims.Roles, Permissions: claims.Permissions, IdentityType: protobufIdentityType(claims.IdentityType)}
}

func membershipAuthorization(record membership.Membership) *identity.EffectiveAuthorization {
	return &identity.EffectiveAuthorization{UserId: record.UserID, MembershipId: record.ID, OrganizationId: record.OrganizationID, Roles: record.Roles, Permissions: record.Permissions, IdentityType: identity.IdentityType_IDENTITY_TYPE_HUMAN}
}

func identityType(value identity.IdentityType) (token.IdentityType, error) {
	switch value {
	case identity.IdentityType_IDENTITY_TYPE_HUMAN:
		return token.IdentityHuman, nil
	case identity.IdentityType_IDENTITY_TYPE_SERVICE:
		return token.IdentityService, nil
	case identity.IdentityType_IDENTITY_TYPE_INTEGRATION:
		return token.IdentityIntegration, nil
	default:
		return "", fmt.Errorf("invalid identity type")
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
