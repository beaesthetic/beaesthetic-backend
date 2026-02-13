package domain

import (
	"errors"
	"time"
)

var (
	ErrConsentNotFound      = errors.New("consent not found")
	ErrConsentAlreadyExists = errors.New("consent already exists for this policy version")
	ErrConsentAlreadyRevoked = errors.New("consent already revoked")
	ErrInvalidSubject       = errors.New("invalid subject")
)

// AcceptanceMethod represents how the consent was accepted
type AcceptanceMethod string

const (
	AcceptanceMethodDirect AcceptanceMethod = "direct"
	AcceptanceMethodLink   AcceptanceMethod = "link"
)

// Consent represents a subject's consent to a specific policy version
type Consent struct {
	ID               string           `bson:"_id" json:"id"`
	TenantID         string           `bson:"tenant_id" json:"tenant_id"`
	Subject          string           `bson:"subject" json:"subject"`
	PolicySlug       string           `bson:"policy_slug" json:"policy_slug"`
	PolicyVersion    string           `bson:"policy_version" json:"policy_version"`
	AcceptedAt       time.Time        `bson:"accepted_at" json:"accepted_at"`
	AcceptanceMethod AcceptanceMethod `bson:"acceptance_method" json:"acceptance_method"`
	LinkToken        *string          `bson:"link_token,omitempty" json:"link_token,omitempty"`
	RevokedAt        *time.Time       `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`
	RevokedBy        *string          `bson:"revoked_by,omitempty" json:"revoked_by,omitempty"`
}

// NewConsent creates a new consent record
func NewConsent(id, tenantID, subject, policySlug, policyVersion string, method AcceptanceMethod, linkToken *string) (*Consent, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}
	if subject == "" {
		return nil, ErrInvalidSubject
	}
	if policySlug == "" {
		return nil, ErrInvalidPolicySlug
	}
	if policyVersion == "" {
		return nil, ErrInvalidPolicyVersion
	}

	return &Consent{
		ID:               id,
		TenantID:         tenantID,
		Subject:          subject,
		PolicySlug:       policySlug,
		PolicyVersion:    policyVersion,
		AcceptedAt:       time.Now().UTC(),
		AcceptanceMethod: method,
		LinkToken:        linkToken,
	}, nil
}

// Revoke marks the consent as revoked
func (c *Consent) Revoke(revokedBy string) error {
	if c.RevokedAt != nil {
		return ErrConsentAlreadyRevoked
	}

	now := time.Now().UTC()
	c.RevokedAt = &now
	c.RevokedBy = &revokedBy
	return nil
}

// IsActive returns true if the consent is not revoked
func (c *Consent) IsActive() bool {
	return c.RevokedAt == nil
}

// SubjectConsents represents all consents for a subject
type SubjectConsents struct {
	Subject  string    `json:"subject"`
	Consents []Consent `json:"consents"`
}

// ConsentRepository defines the interface for consent persistence
type ConsentRepository interface {
	FindByID(id string) (*Consent, error)
	FindBySubject(tenantID, subject string) ([]Consent, error)
	FindBySubjectAndPolicy(tenantID, subject, policySlug string) (*Consent, error)
	FindActiveBySubjectAndPolicy(tenantID, subject, policySlug string) (*Consent, error)
	FindActiveBySubjectAndPolicies(tenantID, subject string, slugs []string) ([]Consent, error)
	Save(consent *Consent) error
	Update(consent *Consent) error
}
