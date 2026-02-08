package application

import (
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
func (s *ConsentService) CreateConsents(req CreateConsentRequest, method domain.AcceptanceMethod, linkToken *string) ([]domain.Consent, error) {
	var consents []domain.Consent

	for _, pc := range req.Policies {
		// Get policy to validate and get version
		policy, err := s.policyRepo.FindBySlug(pc.Slug)
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
		existing, _ := s.consentRepo.FindActiveBySubjectAndPolicy(req.Subject, pc.Slug)
		if existing != nil && existing.PolicyVersion == version {
			// Skip if already consented to this version
			consents = append(consents, *existing)
			continue
		}

		// Create new consent
		consent, err := domain.NewConsent(
			uuid.New().String(),
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
func (s *ConsentService) GetConsentsBySubject(subject string) (*domain.SubjectConsents, error) {
	consents, err := s.consentRepo.FindBySubject(subject)
	if err != nil {
		return nil, err
	}

	return &domain.SubjectConsents{
		Subject:  subject,
		Consents: consents,
	}, nil
}

// GetConsentBySubjectAndPolicy retrieves the consent for a specific policy
func (s *ConsentService) GetConsentBySubjectAndPolicy(subject, policySlug string) (*domain.Consent, error) {
	return s.consentRepo.FindBySubjectAndPolicy(subject, policySlug)
}

// GetActiveConsentBySubjectAndPolicy retrieves the active consent for a specific policy
func (s *ConsentService) GetActiveConsentBySubjectAndPolicy(subject, policySlug string) (*domain.Consent, error) {
	return s.consentRepo.FindActiveBySubjectAndPolicy(subject, policySlug)
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
