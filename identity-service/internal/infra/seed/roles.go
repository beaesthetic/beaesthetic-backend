package seed

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/infra/postgres/queries"
	"gopkg.in/yaml.v3"
)

type Catalog struct { Roles []Role `yaml:"roles"` }
type Role struct { Name string `yaml:"name"`; Permissions []string `yaml:"permissions"` }

func Apply(ctx context.Context, pool *pgxpool.Pool, path string) error {
	body, err := os.ReadFile(path); if err != nil { return err }
	var catalog Catalog; if err := yaml.Unmarshal(body, &catalog); err != nil { return err }
	tx, err := pool.Begin(ctx); if err != nil { return err }; defer tx.Rollback(ctx)
	q := queries.New(tx)
	for _, role := range catalog.Roles {
		roleID, err := q.UpsertGlobalRole(ctx, queries.UpsertGlobalRoleParams{ID: stableID("role:" + role.Name), Name: role.Name}); if err != nil { return err }
		for _, permission := range role.Permissions { permissionID, err := q.UpsertPermission(ctx, queries.UpsertPermissionParams{ID: stableID("permission:" + permission), Code: permission}); if err != nil { return err }; if err := q.GrantRolePermission(ctx, queries.GrantRolePermissionParams{RoleID: roleID, PermissionID: permissionID}); err != nil { return err } }
	}
	return tx.Commit(ctx)
}

func stableID(value string) string { sum := md5.Sum([]byte(value)); return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16]) }
