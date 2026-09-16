package oauth

import (
	"context"

	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/membership"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/token"
)

type MembershipAuthorizer struct{ repository membership.Repository }

func NewMembershipAuthorizer(repository membership.Repository) *MembershipAuthorizer {
	return &MembershipAuthorizer{repository: repository}
}
func (a *MembershipAuthorizer) Name() string { return "membership" }

func (a *MembershipAuthorizer) Authorize(ctx context.Context, identity ExternalIdentity, organizationID string) (token.IssueRequest, error) {
	user, record, err := membership.Resolve(ctx, a.repository, identity.Provider, identity.Subject, identity.Email, organizationID)
	if err != nil {
		return token.IssueRequest{}, err
	}
	return token.IssueRequest{Subject: user.ID, IdentityType: token.IdentityHuman, OrganizationID: record.OrganizationID, MembershipID: record.ID, Roles: record.Roles, Permissions: record.Permissions}, nil
}
