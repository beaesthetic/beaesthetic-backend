package membership

import (
	"context"
	"errors"
	"sort"

	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/user"
)

var ErrNotMember = errors.New("identity is not an active organization member")

type Membership struct {
	ID             string
	UserID         string
	OrganizationID string
	Roles          []string
	Permissions    []string
	Active         bool
}

type Organization struct {
	ID           string
	Name         string
	MembershipID string
	Roles        []string
}

type Member struct {
	UserID       string
	Email        string
	MembershipID string
	Active       bool
	Roles        []string
}

type Repository interface {
	FindActiveMembership(ctx context.Context, userID, organizationID string) (Membership, error)
}

type OrganizationRepository interface {
	ListOrganizations(ctx context.Context, userID string) ([]Organization, error)
}

type ManagementRepository interface {
	ListMembers(ctx context.Context, organizationID string) ([]Member, error)
	AddMember(ctx context.Context, organizationID, email string, roles []string) (Member, error)
	ReplaceMemberRoles(ctx context.Context, organizationID, userID string, roles []string) (Member, error)
	DeactivateMember(ctx context.Context, organizationID, userID string) error
	ListGlobalRoles(ctx context.Context) ([]string, error)
}

func Resolve(ctx context.Context, users user.Repository, memberships Repository, provider, providerSubject, email, organizationID string) (user.User, Membership, error) {
	resolvedUser, err := users.FindOrCreate(ctx, provider, providerSubject, email)
	if err != nil {
		return user.User{}, Membership{}, err
	}
	membership, err := memberships.FindActiveMembership(ctx, resolvedUser.ID, organizationID)
	if err != nil {
		return user.User{}, Membership{}, err
	}
	if !membership.Active {
		return user.User{}, Membership{}, ErrNotMember
	}
	membership.Roles = unique(membership.Roles)
	membership.Permissions = unique(membership.Permissions)
	return resolvedUser, membership, nil
}

func unique(values []string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		if value != "" {
			seen[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
