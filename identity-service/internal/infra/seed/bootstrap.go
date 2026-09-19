package seed

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/infra/postgres/queries"
	"gopkg.in/yaml.v3"
)

type BootstrapCatalog struct {
	Organizations []BootstrapOrganization `yaml:"organizations"`
}

type BootstrapOrganization struct {
	Name  string          `yaml:"name"`
	Users []BootstrapUser `yaml:"users"`
}

type BootstrapUser struct {
	Email string   `yaml:"email"`
	Roles []string `yaml:"roles"`
}

func ApplyBootstrap(ctx context.Context, pool *pgxpool.Pool, path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var catalog BootstrapCatalog
	if err := yaml.Unmarshal(body, &catalog); err != nil {
		return err
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := queries.New(tx)
	for _, organization := range catalog.Organizations {
		if err := validateBootstrapOrganization(organization); err != nil {
			return err
		}
		organizationID, err := q.UpsertOrganization(ctx, queries.UpsertOrganizationParams{
			ID: stableID("organization:" + organization.Name), Name: organization.Name,
		})
		if err != nil {
			return err
		}
		for _, bootstrapUser := range organization.Users {
			userID := stableID("user:email:" + strings.ToLower(strings.TrimSpace(bootstrapUser.Email)))
			if _, err := q.UpsertUser(ctx, queries.UpsertUserParams{ID: userID, Email: bootstrapUser.Email}); err != nil {
				return err
			}
			membershipID, err := q.UpsertMembership(ctx, queries.UpsertMembershipParams{
				ID: stableID("membership:" + userID + ":" + organizationID), UserID: userID, OrganizationID: organizationID,
			})
			if err != nil {
				return err
			}
			for _, role := range bootstrapUser.Roles {
				roleID, err := q.UpsertGlobalRole(ctx, queries.UpsertGlobalRoleParams{ID: stableID("role:" + role), Name: role})
				if err != nil {
					return err
				}
				if err := q.AssignMembershipRole(ctx, queries.AssignMembershipRoleParams{MembershipID: membershipID, RoleID: roleID}); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit(ctx)
}

func validateBootstrapOrganization(organization BootstrapOrganization) error {
	if strings.TrimSpace(organization.Name) == "" {
		return fmt.Errorf("bootstrap organization name is required")
	}
	if len(organization.Users) == 0 {
		return fmt.Errorf("bootstrap users are required for %q", organization.Name)
	}
	for _, user := range organization.Users {
		if strings.TrimSpace(user.Email) == "" || len(user.Roles) == 0 {
			return fmt.Errorf("bootstrap user email and roles are required for %q", organization.Name)
		}
	}
	return nil
}
