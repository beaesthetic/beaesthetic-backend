package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPolicyNotFound        = errors.New("policy not found")
	ErrPolicyAlreadyExists   = errors.New("policy already exists")
	ErrVersionAlreadyExists  = errors.New("version already exists")
	ErrNoActiveVersion       = errors.New("no active version found")
	ErrInvalidPolicySlug     = errors.New("invalid policy slug")
	ErrInvalidPolicyVersion  = errors.New("invalid policy version")
	ErrInvalidTenantID       = errors.New("invalid tenant id")
)

// PolicyVersion represents a specific version of a policy
type PolicyVersion struct {
	Version               string    `bson:"version" json:"version"`
	PublishedAt           time.Time `bson:"published_at" json:"published_at"`
	ContentHTML           string    `bson:"content_html" json:"content_html"`
	ContentMarkdown       string    `bson:"content_markdown" json:"content_markdown"`
	PDFURL                string    `bson:"pdf_url" json:"pdf_url"`
	IsActive              bool      `bson:"is_active" json:"is_active"`
	RequiresReAcceptance  bool      `bson:"requires_re_acceptance" json:"requires_re_acceptance"`
}

// Policy represents a consent policy (e.g., privacy, marketing, cookies)
type Policy struct {
	ID          string          `bson:"_id" json:"id"`
	TenantID    string          `bson:"tenant_id" json:"tenant_id"`
	Slug        string          `bson:"slug" json:"slug"`
	Name        string          `bson:"name" json:"name"`
	Description string          `bson:"description" json:"description"`
	Versions    []PolicyVersion `bson:"versions" json:"versions"`
	CreatedAt   time.Time       `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `bson:"updated_at" json:"updated_at"`
}

// NewPolicy creates a new policy with the given details
func NewPolicy(tenantID, slug, name, description string) (*Policy, error) {
	if tenantID == "" {
		return nil, ErrInvalidTenantID
	}
	if slug == "" {
		return nil, ErrInvalidPolicySlug
	}

	now := time.Now().UTC()
	return &Policy{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Slug:        slug,
		Name:        name,
		Description: description,
		Versions:    []PolicyVersion{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// AddVersion adds a new version to the policy
func (p *Policy) AddVersion(version PolicyVersion) error {
	if version.Version == "" {
		return ErrInvalidPolicyVersion
	}

	// Check if version already exists
	for _, v := range p.Versions {
		if v.Version == version.Version {
			return ErrVersionAlreadyExists
		}
	}

	// Deactivate all other versions if this one is active
	if version.IsActive {
		for i := range p.Versions {
			p.Versions[i].IsActive = false
		}
	}

	version.PublishedAt = time.Now().UTC()
	p.Versions = append(p.Versions, version)
	p.UpdatedAt = time.Now().UTC()

	return nil
}

// GetActiveVersion returns the currently active version of the policy
func (p *Policy) GetActiveVersion() (*PolicyVersion, error) {
	for _, v := range p.Versions {
		if v.IsActive {
			return &v, nil
		}
	}
	return nil, ErrNoActiveVersion
}

// GetVersion returns a specific version of the policy
func (p *Policy) GetVersion(version string) (*PolicyVersion, error) {
	for _, v := range p.Versions {
		if v.Version == version {
			return &v, nil
		}
	}
	return nil, ErrInvalidPolicyVersion
}

// PolicyRepository defines the interface for policy persistence
type PolicyRepository interface {
	FindBySlug(tenantID, slug string) (*Policy, error)
	FindBySlugs(tenantID string, slugs []string) ([]Policy, error)
	FindAll(tenantID string) ([]Policy, error)
	FindAllActive(tenantID string) ([]Policy, error)
	Save(policy *Policy) error
	Update(policy *Policy) error
}
