package application

import (
	"fmt"
	"time"

	"github.com/beaesthetic/consent-service/internal/domain"
	"github.com/google/uuid"
)

// CreateLinkRequest represents the request to create a consent link
type CreateLinkRequest struct {
	Subject        string   `json:"subject" validate:"required"`
	Policies       []string `json:"policies" validate:"required,min=1"`
	ExpiresInHours int      `json:"expires_in_hours" validate:"required,min=1,max=168"` // max 7 days
	CreatedBy      string   `json:"created_by" validate:"required"`
}

// CreateLinkResponse represents the response after creating a consent link
type CreateLinkResponse struct {
	Token     string     `json:"token"`
	URL       string     `json:"url"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// LinkPolicyInfo represents policy information for a link
type LinkPolicyInfo struct {
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Version         string `json:"version"`
	ContentHTML     string `json:"content_html"`
	ContentMarkdown string `json:"content_markdown"`
	PDFURL          string `json:"pdf_url"`
}

// LinkInfo represents the information returned for a consent link
type LinkInfo struct {
	Token     string           `json:"token"`
	Subject   string           `json:"subject"`
	Policies  []LinkPolicyInfo `json:"policies"`
	ExpiresAt time.Time        `json:"expires_at"`
}

// LinkService handles consent link operations
type LinkService struct {
	linkRepo   domain.ConsentLinkRepository
	policyRepo domain.PolicyRepository
	baseURL    string
}

// NewLinkService creates a new LinkService
func NewLinkService(linkRepo domain.ConsentLinkRepository, policyRepo domain.PolicyRepository, baseURL string) *LinkService {
	return &LinkService{
		linkRepo:   linkRepo,
		policyRepo: policyRepo,
		baseURL:    baseURL,
	}
}

// CreateLink creates a new consent link
func (s *LinkService) CreateLink(tenantID string, req CreateLinkRequest) (*CreateLinkResponse, error) {
	// Validate all policies exist and have active versions
	for _, slug := range req.Policies {
		policy, err := s.policyRepo.FindBySlug(tenantID, slug)
		if err != nil {
			return nil, err
		}
		if _, err := policy.GetActiveVersion(); err != nil {
			return nil, fmt.Errorf("policy %s: %w", slug, err)
		}
	}

	token := uuid.New().String()
	link, err := domain.NewConsentLink(token, tenantID, req.Subject, req.Policies, req.ExpiresInHours, req.CreatedBy)
	if err != nil {
		return nil, err
	}

	if err := s.linkRepo.Save(link); err != nil {
		return nil, err
	}

	return &CreateLinkResponse{
		Token:     link.Token,
		URL:       fmt.Sprintf("%s/public/links/%s", s.baseURL, link.Token),
		ExpiresAt: &link.ExpiresAt,
	}, nil
}

// GetLink retrieves link information
func (s *LinkService) GetLink(token string) (*domain.ConsentLink, error) {
	return s.linkRepo.FindByToken(token)
}

// GetLinkInfo retrieves detailed link information including policies
func (s *LinkService) GetLinkInfo(token string) (*LinkInfo, error) {
	link, err := s.linkRepo.FindByToken(token)
	if err != nil {
		return nil, err
	}

	// Validate link is still valid
	if err := link.IsValid(); err != nil {
		return nil, err
	}

	// Get policy details using link's tenant
	var policies []LinkPolicyInfo
	for _, slug := range link.Policies {
		policy, err := s.policyRepo.FindBySlug(link.TenantID, slug)
		if err != nil {
			return nil, err
		}

		activeVersion, err := policy.GetActiveVersion()
		if err != nil {
			return nil, fmt.Errorf("policy %s: %w", slug, err)
		}

		policies = append(policies, LinkPolicyInfo{
			Slug:            policy.Slug,
			Name:            policy.Name,
			Description:     policy.Description,
			Version:         activeVersion.Version,
			ContentHTML:     activeVersion.ContentHTML,
			ContentMarkdown: activeVersion.ContentMarkdown,
			PDFURL:          activeVersion.PDFURL,
		})
	}

	return &LinkInfo{
		Token:     link.Token,
		Subject:   link.Subject,
		Policies:  policies,
		ExpiresAt: link.ExpiresAt,
	}, nil
}

// InvalidateLink invalidates a consent link
func (s *LinkService) InvalidateLink(token string) error {
	return s.linkRepo.Delete(token)
}

// MarkLinkAsUsed marks a link as used
func (s *LinkService) MarkLinkAsUsed(token string) error {
	link, err := s.linkRepo.FindByToken(token)
	if err != nil {
		return err
	}

	if err := link.MarkAsUsed(); err != nil {
		return err
	}

	return s.linkRepo.Update(link)
}
