package application

import (
	"github.com/beaesthetic/consent-service/internal/domain"
)

// CreatePolicyRequest represents the request to create a new policy
type CreatePolicyRequest struct {
	Slug        string `json:"slug" validate:"required,min=1,max=100"`
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description" validate:"max=1000"`
}

// AddVersionRequest represents the request to add a new version to a policy
type AddVersionRequest struct {
	Version         string `json:"version" validate:"required,min=1,max=20"`
	ContentHTML     string `json:"content_html"`
	ContentMarkdown string `json:"content_markdown"`
	PDFURL          string `json:"pdf_url" validate:"omitempty,url"`
	IsActive        bool   `json:"is_active"`
}

// PolicyService handles policy-related operations
type PolicyService struct {
	policyRepo domain.PolicyRepository
}

// NewPolicyService creates a new PolicyService
func NewPolicyService(policyRepo domain.PolicyRepository) *PolicyService {
	return &PolicyService{
		policyRepo: policyRepo,
	}
}

// CreatePolicy creates a new policy
func (s *PolicyService) CreatePolicy(req CreatePolicyRequest) (*domain.Policy, error) {
	// Check if policy already exists
	existing, _ := s.policyRepo.FindBySlug(req.Slug)
	if existing != nil {
		return nil, domain.ErrPolicyAlreadyExists
	}

	policy, err := domain.NewPolicy(req.Slug, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	if err := s.policyRepo.Save(policy); err != nil {
		return nil, err
	}

	return policy, nil
}

// GetPolicy retrieves a policy by slug
func (s *PolicyService) GetPolicy(slug string) (*domain.Policy, error) {
	return s.policyRepo.FindBySlug(slug)
}

// GetAllPolicies retrieves all policies
func (s *PolicyService) GetAllPolicies() ([]domain.Policy, error) {
	return s.policyRepo.FindAll()
}

// GetActivePolicies retrieves all policies with at least one active version
func (s *PolicyService) GetActivePolicies() ([]domain.Policy, error) {
	return s.policyRepo.FindAllActive()
}

// AddVersion adds a new version to an existing policy
func (s *PolicyService) AddVersion(slug string, req AddVersionRequest) (*domain.Policy, error) {
	policy, err := s.policyRepo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	version := domain.PolicyVersion{
		Version:         req.Version,
		ContentHTML:     req.ContentHTML,
		ContentMarkdown: req.ContentMarkdown,
		PDFURL:          req.PDFURL,
		IsActive:        req.IsActive,
	}

	if err := policy.AddVersion(version); err != nil {
		return nil, err
	}

	if err := s.policyRepo.Update(policy); err != nil {
		return nil, err
	}

	return policy, nil
}

// GetActiveVersion retrieves the active version of a policy
func (s *PolicyService) GetActiveVersion(slug string) (*domain.PolicyVersion, error) {
	policy, err := s.policyRepo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	return policy.GetActiveVersion()
}
