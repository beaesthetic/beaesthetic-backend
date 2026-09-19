package postgres

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/domain/membership"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/infra/postgres/queries"
)

type MembershipRepository struct{ pool *pgxpool.Pool }

func NewMembershipRepository(pool *pgxpool.Pool) *MembershipRepository {
	return &MembershipRepository{pool: pool}
}

func (r *MembershipRepository) FindOrCreateUser(ctx context.Context, provider, providerSubject, email string) (membership.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return membership.User{}, err
	}
	defer tx.Rollback(ctx)
	var user membership.User
	err = tx.QueryRow(ctx, `SELECT u.id::text, coalesce(u.email, '') FROM external_identities e JOIN users u ON u.id = e.user_id WHERE e.provider = $1 AND e.subject = $2`, provider, providerSubject).Scan(&user.ID, &user.Email)
	if err == nil {
		user.Provider, user.ProviderSubject = provider, providerSubject
		return user, tx.Commit(ctx)
	}
	if err != pgx.ErrNoRows {
		return membership.User{}, err
	}
	// Bootstrap users are created by email before their Firebase subject exists.
	// Link the first authenticated external identity to that pre-provisioned user.
	if email != "" {
		q := queries.New(tx)
		byEmail, emailErr := q.FindUserByEmail(ctx, email)
		if emailErr == nil {
			if err := q.CreateExternalIdentity(ctx, queries.CreateExternalIdentityParams{Provider: provider, Subject: providerSubject, UserID: byEmail.ID}); err != nil {
				return membership.User{}, err
			}
			user = membership.User{ID: byEmail.ID, Email: byEmail.Email, Provider: provider, ProviderSubject: providerSubject}
			return user, tx.Commit(ctx)
		}
		if emailErr != pgx.ErrNoRows {
			return membership.User{}, emailErr
		}
	}
	user.ID, err = newID()
	if err != nil {
		return membership.User{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, NULLIF($2, ''))`, user.ID, email); err != nil {
		return membership.User{}, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO external_identities (provider, subject, user_id) VALUES ($1, $2, $3)`, provider, providerSubject, user.ID); err != nil {
		return membership.User{}, err
	}
	user.Provider, user.ProviderSubject, user.Email = provider, providerSubject, email
	return user, tx.Commit(ctx)
}

func (r *MembershipRepository) FindActiveMembership(ctx context.Context, userID, organizationID string) (membership.Membership, error) {
	var membershipRecord membership.Membership
	err := r.pool.QueryRow(ctx, `SELECT id::text, user_id::text, organization_id::text, active FROM memberships WHERE user_id = $1 AND organization_id = $2`, userID, organizationID).Scan(&membershipRecord.ID, &membershipRecord.UserID, &membershipRecord.OrganizationID, &membershipRecord.Active)
	if err != nil {
		return membership.Membership{}, err
	}
	roles, err := r.pool.Query(ctx, `SELECT r.name FROM membership_roles mr JOIN roles r ON r.id = mr.role_id WHERE mr.membership_id = $1 ORDER BY r.name`, membershipRecord.ID)
	if err != nil {
		return membership.Membership{}, err
	}
	defer roles.Close()
	for roles.Next() {
		var role string
		if err := roles.Scan(&role); err != nil {
			return membership.Membership{}, err
		}
		membershipRecord.Roles = append(membershipRecord.Roles, role)
	}
	if err := roles.Err(); err != nil {
		return membership.Membership{}, err
	}
	permissions, err := r.pool.Query(ctx, `SELECT DISTINCT p.code FROM membership_roles mr JOIN role_permissions rp ON rp.role_id = mr.role_id JOIN permissions p ON p.id = rp.permission_id WHERE mr.membership_id = $1 ORDER BY p.code`, membershipRecord.ID)
	if err != nil {
		return membership.Membership{}, err
	}
	defer permissions.Close()
	for permissions.Next() {
		var permission string
		if err := permissions.Scan(&permission); err != nil {
			return membership.Membership{}, err
		}
		membershipRecord.Permissions = append(membershipRecord.Permissions, permission)
	}
	return membershipRecord, permissions.Err()
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
