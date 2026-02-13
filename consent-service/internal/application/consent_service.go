package application

import (
	"errors"
	"fmt"

	"github.com/beaesthetic/consent-service/internal/domain"
	"github.com/google/uuid"
)

// PolicyConsent represents a policy to consent to
type PolicyConsent struct {
	Slug    string `json:"slug" validate:"required"`
	Version string `json:"version,omitempty"` // If empty, uses active version
}

// CreateConsentRequest represents the request to create consents
type CreateConsentRequest struct {
	Subject  string          `json:"subject" validate:"required"`
	Policies []PolicyConsent `json:"policies" validate:"required,min=1,dive"`
}

// RevokeConsentRequest represents the request to revoke a consent
type RevokeConsentRequest struct {
	RevokedBy string `json:"revoked_by" validate:"required"`
}

// ConsentService handles consent-related operations
type ConsentService struct {
	consentRepo domain.ConsentRepository
	policyRepo  domain.PolicyRepository
}

// NewConsentService creates a new ConsentService
func NewConsentService(consentRepo domain.ConsentRepository, policyRepo domain.PolicyRepository) *ConsentService {
	return &ConsentService{
		consentRepo: consentRepo,
		policyRepo:  policyRepo,
	}
}

// CreateConsents creates consent records for the given policies
func (s *ConsentService) CreateConsents(tenantID string, req CreateConsentRequest, method domain.AcceptanceMethod, linkToken *string) ([]domain.Consent, error) {
	var consents []domain.Consent

	for _, pc := range req.Policies {
		// Get policy to validate and get version
		policy, err := s.policyRepo.FindBySlug(tenantID, pc.Slug)
		if err != nil {
			return nil, err
		}

		// Determine version
		version := pc.Version
		if version == "" {
			activeVersion, err := policy.GetActiveVersion()
			if err != nil {
				return nil, err
			}
			version = activeVersion.Version
		} else {
			// Validate version exists
			if _, err := policy.GetVersion(version); err != nil {
				return nil, err
			}
		}

		// Check if consent already exists (active)
		existing, _ := s.consentRepo.FindActiveBySubjectAndPolicy(tenantID, req.Subject, pc.Slug)
		if existing != nil && existing.PolicyVersion == version {
			// Skip if already consented to this version
			consents = append(consents, *existing)
			continue
		}

		// Create new consent
		consent, err := domain.NewConsent(
			uuid.New().String(),
			tenantID,
			req.Subject,
			pc.Slug,
			version,
			method,
			linkToken,
		)
		if err != nil {
			return nil, err
		}

		if err := s.consentRepo.Save(consent); err != nil {
			return nil, err
		}

		consents = append(consents, *consent)
	}

	return consents, nil
}

// GetConsentsBySubject retrieves all consents for a subject
func (s *ConsentService) GetConsentsBySubject(tenantID, subject string) (*domain.SubjectConsents, error) {
	consents, err := s.consentRepo.FindBySubject(tenantID, subject)
	if err != nil {
		return nil, err
	}

	return &domain.SubjectConsents{
		Subject:  subject,
		Consents: consents,
	}, nil
}

// GetConsentBySubjectAndPolicy retrieves the consent for a specific policy
func (s *ConsentService) GetConsentBySubjectAndPolicy(tenantID, subject, policySlug string) (*domain.Consent, error) {
	return s.consentRepo.FindBySubjectAndPolicy(tenantID, subject, policySlug)
}

// GetActiveConsentBySubjectAndPolicy retrieves the active consent for a specific policy
func (s *ConsentService) GetActiveConsentBySubjectAndPolicy(tenantID, subject, policySlug string) (*domain.Consent, error) {
	return s.consentRepo.FindActiveBySubjectAndPolicy(tenantID, subject, policySlug)
}

// GetConsentByID retrieves a consent by ID
func (s *ConsentService) GetConsentByID(id string) (*domain.Consent, error) {
	return s.consentRepo.FindByID(id)
}

// RevokeConsent revokes a consent
func (s *ConsentService) RevokeConsent(id string, revokedBy string) (*domain.Consent, error) {
	consent, err := s.consentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if err := consent.Revoke(revokedBy); err != nil {
		return nil, err
	}

	if err := s.consentRepo.Update(consent); err != nil {
		return nil, err
	}

	return consent, nil
}

// GetConsentStatus returns the consent status for a subject and a list of policy slugs.
// If slugs is empty, all active policies of the tenant are used.
func (s *ConsentService) GetConsentStatus(tenantID, subject string, slugs []string) ([]domain.PolicyConsentStatus, error) {
	if tenantID == "" {
		return nil, domain.ErrInvalidTenantID
	}
	if subject == "" {
		return nil, domain.ErrInvalidSubject
	}

	var policies []domain.Policy
	var err error

	if len(slugs) == 0 {
		// No slugs provided: use all active policies of the tenant
		policies, err = s.policyRepo.FindAllActive(tenantID)
		if err != nil {
			return nil, err
		}
	} else {
		// Fetch requested policies
		policies, err = s.policyRepo.FindBySlugs(tenantID, slugs)
		if err != nil {
			return nil, err
		}

		// Check all slugs were found
		found := make(map[string]bool, len(policies))
		for _, p := range policies {
			found[p.Slug] = true
		}
		var notFound []string
		for _, slug := range slugs {
			if !found[slug] {
				notFound = append(notFound, slug)
			}
		}
		if len(notFound) > 0 {
			return nil, fmt.Errorf("policies not found: %v: %w", notFound, domain.ErrPolicyNotFound)
		}
	}

	// Build slug list from policies
	policySlugs := make([]string, len(policies))
	policyMap := make(map[string]domain.Policy, len(policies))
	for i, p := range policies {
		policySlugs[i] = p.Slug
		policyMap[p.Slug] = p
	}

	// Fetch all active consents for subject+policies in one query
	consents, err := s.consentRepo.FindActiveBySubjectAndPolicies(tenantID, subject, policySlugs)
	if err != nil && !errors.Is(err, domain.ErrConsentNotFound) {
		return nil, err
	}

	// Build consent map: slug → most recent active consent
	consentMap := make(map[string]*domain.Consent, len(consents))
	for i, c := range consents {
		if _, exists := consentMap[c.PolicySlug]; !exists {
			consentMap[c.PolicySlug] = &consents[i]
		}
	}

	// Compute status for each policy
	var statuses []domain.PolicyConsentStatus
	for _, slug := range policySlugs {
		policy := policyMap[slug]

		activeVersion, err := policy.GetActiveVersion()
		if err != nil {
			// Policy without active version: skip
			continue
		}

		consent := consentMap[slug]
		status := domain.ComputeConsentStatus(policy, *activeVersion, consent)
		statuses = append(statuses, status)
	}

	return statuses, nil
}
