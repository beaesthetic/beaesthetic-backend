package application

import (
	"context"

	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/membership"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/token"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/user"
)

type MembershipAuthorizer struct {
	users       user.Repository
	memberships membership.Repository
}

func NewMembershipAuthorizer(users user.Repository, memberships membership.Repository) *MembershipAuthorizer {
	return &MembershipAuthorizer{users: users, memberships: memberships}
}
func (a *MembershipAuthorizer) Name() string { return "membership" }

func (a *MembershipAuthorizer) Authorize(ctx context.Context, identity ExternalIdentity, organizationID string) (token.IssueRequest, error) {
	user, record, err := membership.Resolve(ctx, a.users, a.memberships, identity.Provider, identity.Subject, identity.Email, organizationID)
	if err != nil {
		return token.IssueRequest{}, err
	}
	return token.IssueRequest{
		Subject:        user.ID,
		IdentityType:   token.IdentityHuman,
		Scope:          token.ScopeMembership,
		OrganizationID: record.OrganizationID,
		MembershipID:   record.ID,
		Roles:          record.Roles,
		Permissions:    record.Permissions,
	}, nil
}
