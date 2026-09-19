-- name: FindUserByExternalIdentity :one
SELECT u.id::text, coalesce(u.email,'') AS email FROM external_identities e JOIN users u ON u.id=e.user_id WHERE e.provider=$1 AND e.subject=$2;
-- name: CreateUser :exec
INSERT INTO users (id,email) VALUES ($1,NULLIF($2,''));
-- name: CreateExternalIdentity :exec
INSERT INTO external_identities (provider,subject,user_id) VALUES ($1,$2,$3);
-- name: FindMembership :one
SELECT id::text,user_id::text,organization_id::text,active FROM memberships WHERE user_id=$1 AND organization_id=$2;
-- name: ListMembershipRoles :many
SELECT r.name FROM membership_roles mr JOIN roles r ON r.id=mr.role_id WHERE mr.membership_id=$1 ORDER BY r.name;
-- name: ListMembershipPermissions :many
SELECT DISTINCT p.code FROM membership_roles mr JOIN role_permissions rp ON rp.role_id=mr.role_id JOIN permissions p ON p.id=rp.permission_id WHERE mr.membership_id=$1 ORDER BY p.code;
-- name: UpsertGlobalRole :one
INSERT INTO roles (id, organization_id, name) VALUES ($1, NULL, $2) ON CONFLICT (organization_id, name) DO UPDATE SET name = EXCLUDED.name RETURNING id::text;
-- name: UpsertPermission :one
INSERT INTO permissions (id, code) VALUES ($1, $2) ON CONFLICT (code) DO UPDATE SET code = EXCLUDED.code RETURNING id::text;
-- name: GrantRolePermission :exec
INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;
