package token

import "github.com/petretiandrea/beaesthetic-backend/core/auth"

type IssueRequest struct {
	Subject        string            `json:"subject"`
	IdentityType   auth.IdentityType `json:"identity_type"`
	Scope          auth.Scope        `json:"token_scope"`
	OrganizationID string            `json:"organization_id,omitempty"`
	MembershipID   string            `json:"membership_id,omitempty"`
	Roles          []string          `json:"roles,omitempty"`
	Permissions    []string          `json:"permissions"`
	AuthMethod     string            `json:"auth_method"`
}
