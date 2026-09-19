package application

import (
	"context"
	"fmt"

	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/token"
)

const TokenExchangeGrant = "urn:ietf:params:oauth:grant-type:token-exchange"
const JWTSubjectTokenType = "urn:ietf:params:oauth:token-type:jwt"

type ExternalIdentity struct{ Provider, Subject, Email string }

// Authenticator proves an external assertion and returns a normalized identity.
// Firebase is only one implementation; OIDC, mTLS, or partner adapters can be
// registered without changing the exchange flow.
type Authenticator interface {
	Name() string
	Authenticate(ctx context.Context, assertion string) (ExternalIdentity, error)
}

// Authorizer grants an authenticated identity a local authorization context.
// It must resolve permissions itself: callers never submit roles or permissions.
type Authorizer interface {
	Name() string
	Authorize(ctx context.Context, identity ExternalIdentity, organizationID string) (token.IssueRequest, error)
}

type ExchangeService struct {
	issuer         *token.Issuer
	authenticators map[string]Authenticator
	authorizers    map[string]Authorizer
}

func NewExchangeService(issuer *token.Issuer, authenticators []Authenticator, authorizers []Authorizer) *ExchangeService {
	registeredAuthenticators := make(map[string]Authenticator, len(authenticators))
	for _, authenticator := range authenticators {
		registeredAuthenticators[authenticator.Name()] = authenticator
	}
	registeredAuthorizers := make(map[string]Authorizer, len(authorizers))
	for _, authorizer := range authorizers {
		registeredAuthorizers[authorizer.Name()] = authorizer
	}
	return &ExchangeService{issuer: issuer, authenticators: registeredAuthenticators, authorizers: registeredAuthorizers}
}

type Request struct{ GrantType, SubjectTokenType, Authenticator, Authorizer, SubjectToken, Audience, OrganizationID string }

func (s *ExchangeService) Exchange(ctx context.Context, request Request) (string, token.Claims, error) {
	if request.GrantType != TokenExchangeGrant || request.SubjectTokenType != JWTSubjectTokenType || request.Authenticator == "" || request.Authorizer == "" || request.SubjectToken == "" || request.Audience == "" || request.OrganizationID == "" {
		return "", token.Claims{}, fmt.Errorf("invalid token exchange request")
	}
	authenticator, ok := s.authenticators[request.Authenticator]
	if !ok {
		return "", token.Claims{}, fmt.Errorf("unsupported authenticator %q", request.Authenticator)
	}
	identity, err := authenticator.Authenticate(ctx, request.SubjectToken)
	if err != nil {
		return "", token.Claims{}, fmt.Errorf("verify external token: %w", err)
	}
	authorizer, ok := s.authorizers[request.Authorizer]
	if !ok {
		return "", token.Claims{}, fmt.Errorf("unsupported authorizer %q", request.Authorizer)
	}
	issueRequest, err := authorizer.Authorize(ctx, identity, request.OrganizationID)
	if err != nil {
		return "", token.Claims{}, fmt.Errorf("resolve authorization: %w", err)
	}
	issueRequest.Audience = request.Audience
	issueRequest.AuthMethod = identity.Provider
	return s.issuer.Issue(issueRequest)
}
