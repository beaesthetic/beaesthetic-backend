package token

import "time"

type IdentityType string

const (
	IdentityHuman       IdentityType = "human"
	IdentityService     IdentityType = "service"
	IdentityIntegration IdentityType = "integration"
)

func (t IdentityType) Valid() bool {
	return t == IdentityHuman || t == IdentityService || t == IdentityIntegration
}

type Scope string

const (
	ScopeUser       Scope = "user"
	ScopeMembership Scope = "membership"
)

func (s Scope) Valid() bool { return s == ScopeUser || s == ScopeMembership }

type IssueRequest struct {
	Subject        string       `json:"subject"`
	IdentityType   IdentityType `json:"identity_type"`
	Scope          Scope        `json:"token_scope"`
	OrganizationID string       `json:"organization_id,omitempty"`
	MembershipID   string       `json:"membership_id,omitempty"`
	Roles          []string     `json:"roles,omitempty"`
	Permissions    []string     `json:"permissions"`
	AuthMethod     string       `json:"auth_method"`
}

type Claims struct {
	Issuer         string       `json:"iss"`
	Subject        string       `json:"sub"`
	IssuedAt       time.Time    `json:"iat"`
	ExpiresAt      time.Time    `json:"exp"`
	TokenID        string       `json:"jti"`
	KeyID          string       `json:"kid"`
	IdentityType   IdentityType `json:"identity_type"`
	Scope          Scope        `json:"token_scope"`
	OrganizationID string       `json:"organization_id,omitempty"`
	MembershipID   string       `json:"membership_id,omitempty"`
	Roles          []string     `json:"roles,omitempty"`
	Permissions    []string     `json:"permissions"`
	AuthMethod     string       `json:"auth_method"`
}
