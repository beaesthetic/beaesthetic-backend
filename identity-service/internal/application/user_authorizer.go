package application

import (
	"context"
	"fmt"

	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/token"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/user"
)

type UserAuthorizer struct{ users user.Repository }

func NewUserAuthorizer(users user.Repository) *UserAuthorizer {
	return &UserAuthorizer{users: users}
}

func (a *UserAuthorizer) Name() string { return "user" }

func (a *UserAuthorizer) Authorize(ctx context.Context, identity ExternalIdentity, organizationID string) (token.IssueRequest, error) {
	if organizationID != "" {
		return token.IssueRequest{}, fmt.Errorf("user authorizer does not accept organization ID")
	}
	resolvedUser, err := a.users.FindOrCreate(ctx, identity.Provider, identity.Subject, identity.Email)
	if err != nil {
		return token.IssueRequest{}, err
	}
	return token.IssueRequest{
		Subject:      resolvedUser.ID,
		IdentityType: token.IdentityHuman,
		Scope:        token.ScopeUser,
	}, nil
}
