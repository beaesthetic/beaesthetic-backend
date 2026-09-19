package membership

import (
	"context"
	"errors"
	"sort"
)

var ErrNotMember = errors.New("identity is not an active organization member")

type User struct {
	ID              string
	Provider        string
	ProviderSubject string
	Email           string
}

type Membership struct {
	ID             string
	UserID         string
	OrganizationID string
	Roles          []string
	Permissions    []string
	Active         bool
}

// Repository intentionally exposes only authorization reads to the exchange flow.
// Administration of users, organizations, roles and grants belongs behind a separate API.
type Repository interface {
	FindOrCreateUser(ctx context.Context, provider, providerSubject, email string) (User, error)
	FindActiveMembership(ctx context.Context, userID, organizationID string) (Membership, error)
}

func Resolve(ctx context.Context, repository Repository, provider, providerSubject, email, organizationID string) (User, Membership, error) {
	user, err := repository.FindOrCreateUser(ctx, provider, providerSubject, email)
	if err != nil {
		return User{}, Membership{}, err
	}
	membership, err := repository.FindActiveMembership(ctx, user.ID, organizationID)
	if err != nil {
		return User{}, Membership{}, err
	}
	if !membership.Active {
		return User{}, Membership{}, ErrNotMember
	}
	membership.Roles = unique(membership.Roles)
	membership.Permissions = unique(membership.Permissions)
	return user, membership, nil
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
