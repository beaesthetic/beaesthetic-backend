package postgres

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/membership"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/user"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/infra/postgres/queries"
)

type MembershipRepository struct{ pool *pgxpool.Pool }

func NewMembershipRepository(pool *pgxpool.Pool) *MembershipRepository {
	return &MembershipRepository{pool: pool}
}

func (r *MembershipRepository) FindOrCreate(ctx context.Context, provider, providerSubject, email string) (user.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return user.User{}, err
	}
	defer tx.Rollback(ctx)
	q := queries.New(tx)
	resolvedUserRow, err := q.FindUserByExternalIdentity(ctx, queries.FindUserByExternalIdentityParams{
		Provider: provider,
		Subject:  providerSubject,
	})
	if err == nil {
		resolvedUser := user.User{
			ID:              resolvedUserRow.ID,
			Email:           resolvedUserRow.Email,
			Provider:        provider,
			ProviderSubject: providerSubject,
		}
		return resolvedUser, tx.Commit(ctx)
	}
	if err != pgx.ErrNoRows {
		return user.User{}, err
	}
	// Bootstrap users are created by email before their Firebase subject exists.
	// Link the first authenticated external identity to that pre-provisioned user.
	if email != "" {
		byEmail, emailErr := q.FindUserByEmail(ctx, email)
		if emailErr == nil {
			if err := q.CreateExternalIdentity(ctx, queries.CreateExternalIdentityParams{Provider: provider, Subject: providerSubject, UserID: byEmail.ID}); err != nil {
				return user.User{}, err
			}
			resolvedUser := user.User{
				ID:              byEmail.ID,
				Email:           byEmail.Email,
				Provider:        provider,
				ProviderSubject: providerSubject,
			}
			return resolvedUser, tx.Commit(ctx)
		}
		if emailErr != pgx.ErrNoRows {
			return user.User{}, emailErr
		}
	}
	userID, err := newID()
	if err != nil {
		return user.User{}, err
	}
	if err := q.CreateUser(ctx, queries.CreateUserParams{
		ID:    userID,
		Email: email,
	}); err != nil {
		return user.User{}, err
	}
	if err := q.CreateExternalIdentity(ctx, queries.CreateExternalIdentityParams{
		Provider: provider,
		Subject:  providerSubject,
		UserID:   userID,
	}); err != nil {
		return user.User{}, err
	}
	resolvedUser := user.User{
		ID:              userID,
		Provider:        provider,
		ProviderSubject: providerSubject,
		Email:           email,
	}
	return resolvedUser, tx.Commit(ctx)
}

func (r *MembershipRepository) FindActiveMembership(ctx context.Context, userID, organizationID string) (membership.Membership, error) {
	q := queries.New(r.pool)
	record, err := q.FindMembership(ctx, queries.FindMembershipParams{
		UserID:         userID,
		OrganizationID: organizationID,
	})
	if err != nil {
		return membership.Membership{}, err
	}
	membershipRecord := membership.Membership{
		ID:             record.ID,
		UserID:         record.UserID,
		OrganizationID: record.OrganizationID,
		Active:         record.Active,
	}
	roles, err := q.ListMembershipRoles(ctx, membershipRecord.ID)
	if err != nil {
		return membership.Membership{}, err
	}
	membershipRecord.Roles = roles
	permissions, err := q.ListMembershipPermissions(ctx, membershipRecord.ID)
	if err != nil {
		return membership.Membership{}, err
	}
	membershipRecord.Permissions = permissions
	return membershipRecord, nil
}

func (r *MembershipRepository) ListOrganizations(ctx context.Context, userID string) ([]membership.Organization, error) {
	q := queries.New(r.pool)
	rows, err := q.ListOrganizationsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	organizations := make([]membership.Organization, 0, len(rows))
	for _, row := range rows {
		roles, err := q.ListMembershipRoles(ctx, row.MembershipID)
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, membership.Organization{
			ID:           row.OrganizationID,
			Name:         row.OrganizationName,
			MembershipID: row.MembershipID,
			Roles:        roles,
		})
	}
	return organizations, nil
}

func (r *MembershipRepository) ListMembers(ctx context.Context, organizationID string) ([]membership.Member, error) {
	q := queries.New(r.pool)
	rows, err := q.ListOrganizationMembers(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	result := make([]membership.Member, 0, len(rows))
	for _, row := range rows {
		roles, err := q.ListMembershipRoles(ctx, row.MembershipID)
		if err != nil {
			return nil, err
		}
		result = append(result, membership.Member{
			UserID:       row.UserID,
			Email:        row.Email,
			MembershipID: row.MembershipID,
			Active:       row.Active,
			Roles:        roles,
		})
	}
	return result, nil
}

func (r *MembershipRepository) AddMember(ctx context.Context, organizationID, email string, roles []string) (membership.Member, error) {
	if email = strings.TrimSpace(email); email == "" {
		return membership.Member{}, fmt.Errorf("email is required")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return membership.Member{}, err
	}
	defer tx.Rollback(ctx)
	q := queries.New(tx)
	user, err := q.FindUserByEmail(ctx, email)
	if err == pgx.ErrNoRows {
		id, idErr := newID()
		if idErr != nil {
			return membership.Member{}, idErr
		}
		if err = q.CreateUser(ctx, queries.CreateUserParams{
			ID:    id,
			Email: email,
		}); err != nil {
			return membership.Member{}, err
		}
		user.ID, user.Email = id, email
	} else if err != nil {
		return membership.Member{}, err
	}
	membershipIDValue, err := newID()
	if err != nil {
		return membership.Member{}, err
	}
	membershipID, err := q.UpsertMembership(ctx, queries.UpsertMembershipParams{
		ID:             membershipIDValue,
		UserID:         user.ID,
		OrganizationID: organizationID,
	})
	if err != nil {
		return membership.Member{}, err
	}
	if err = replaceRoles(ctx, q, membershipID, roles); err != nil {
		return membership.Member{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return membership.Member{}, err
	}
	return membership.Member{UserID: user.ID, Email: user.Email, MembershipID: membershipID, Active: true, Roles: roles}, nil
}

func (r *MembershipRepository) ReplaceMemberRoles(ctx context.Context, organizationID, userID string, roles []string) (membership.Member, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return membership.Member{}, err
	}
	defer tx.Rollback(ctx)
	q := queries.New(tx)
	record, err := q.FindMembership(ctx, queries.FindMembershipParams{UserID: userID, OrganizationID: organizationID})
	if err != nil {
		return membership.Member{}, err
	}
	if err = replaceRoles(ctx, q, record.ID, roles); err != nil {
		return membership.Member{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return membership.Member{}, err
	}
	return membership.Member{UserID: userID, MembershipID: record.ID, Active: record.Active, Roles: roles}, nil
}

func (r *MembershipRepository) RemoveMember(ctx context.Context, organizationID, userID string) error {
	q := queries.New(r.pool)
	count, err := q.RemoveMembership(ctx, queries.RemoveMembershipParams{UserID: userID, OrganizationID: organizationID})
	if err != nil {
		return err
	}
	if count == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *MembershipRepository) ListGlobalRoles(ctx context.Context) ([]string, error) {
	return queries.New(r.pool).ListGlobalRoles(ctx)
}

func replaceRoles(ctx context.Context, q *queries.Queries, membershipID string, roles []string) error {
	if len(roles) == 0 {
		return fmt.Errorf("at least one role is required")
	}
	if err := q.DeleteMembershipRoles(ctx, membershipID); err != nil {
		return err
	}
	for _, name := range roles {
		role, err := q.FindGlobalRole(ctx, name)
		if err != nil {
			return err
		}
		if err := q.AssignMembershipRole(ctx, queries.AssignMembershipRoleParams{MembershipID: membershipID, RoleID: role.ID}); err != nil {
			return err
		}
	}
	return nil
}

func newID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
