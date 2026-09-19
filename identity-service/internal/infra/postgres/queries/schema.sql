CREATE TABLE users (id UUID PRIMARY KEY, email TEXT);
CREATE TABLE external_identities (provider TEXT NOT NULL, subject TEXT NOT NULL, user_id UUID NOT NULL REFERENCES users(id), PRIMARY KEY (provider, subject));
CREATE TABLE organizations (id UUID PRIMARY KEY, name TEXT NOT NULL);
CREATE TABLE roles (id UUID PRIMARY KEY, organization_id UUID REFERENCES organizations(id), name TEXT NOT NULL, UNIQUE NULLS NOT DISTINCT (organization_id,name));
CREATE TABLE permissions (id UUID PRIMARY KEY, code TEXT NOT NULL UNIQUE);
CREATE TABLE role_permissions (role_id UUID NOT NULL REFERENCES roles(id), permission_id UUID NOT NULL REFERENCES permissions(id), PRIMARY KEY(role_id,permission_id));
CREATE TABLE memberships (id UUID PRIMARY KEY, user_id UUID NOT NULL REFERENCES users(id), organization_id UUID NOT NULL REFERENCES organizations(id), active BOOLEAN NOT NULL DEFAULT true, UNIQUE(user_id,organization_id));
CREATE TABLE membership_roles (membership_id UUID NOT NULL REFERENCES memberships(id), role_id UUID NOT NULL REFERENCES roles(id), PRIMARY KEY(membership_id,role_id));
