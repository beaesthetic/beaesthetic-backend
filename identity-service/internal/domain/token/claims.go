package token

import "github.com/petretiandrea/beaesthetic-backend/core/auth"

// Aliases keep the issuer API tied to the shared access-token contract.
type IdentityType = auth.IdentityType
type Scope = auth.Scope
type Claims = auth.Claims

const (
	IdentityHuman       = auth.IdentityHuman
	IdentityService     = auth.IdentityService
	IdentityIntegration = auth.IdentityIntegration
	ScopeUser           = auth.ScopeUser
	ScopeMembership     = auth.ScopeMembership
)

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
