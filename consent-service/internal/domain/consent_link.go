package domain

import (
	"errors"
	"time"
)

var (
	ErrLinkNotFound     = errors.New("consent link not found")
	ErrLinkExpired      = errors.New("consent link has expired")
	ErrLinkAlreadyUsed  = errors.New("consent link has already been used")
	ErrInvalidLinkToken = errors.New("invalid link token")
)

// ConsentLink represents a shareable link for consent collection
type ConsentLink struct {
	Token     string    `bson:"_id" json:"token"`
	Subject   string    `bson:"subject" json:"subject"`
	Policies  []string  `bson:"policies" json:"policies"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	ExpiresAt time.Time `bson:"expires_at" json:"expires_at"`
	UsedAt    *time.Time `bson:"used_at,omitempty" json:"used_at,omitempty"`
	CreatedBy string    `bson:"created_by" json:"created_by"`
}

// NewConsentLink creates a new consent link
func NewConsentLink(token, subject string, policies []string, expiresInHours int, createdBy string) (*ConsentLink, error) {
	if token == "" {
		return nil, ErrInvalidLinkToken
	}
	if subject == "" {
		return nil, ErrInvalidSubject
	}
	if len(policies) == 0 {
		return nil, ErrInvalidPolicySlug
	}

	now := time.Now().UTC()
	return &ConsentLink{
		Token:     token,
		Subject:   subject,
		Policies:  policies,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(expiresInHours) * time.Hour),
		CreatedBy: createdBy,
	}, nil
}

// IsValid checks if the link is still valid (not expired and not used)
func (l *ConsentLink) IsValid() error {
	if l.UsedAt != nil {
		return ErrLinkAlreadyUsed
	}
	if time.Now().UTC().After(l.ExpiresAt) {
		return ErrLinkExpired
	}
	return nil
}

// MarkAsUsed marks the link as used
func (l *ConsentLink) MarkAsUsed() error {
	if err := l.IsValid(); err != nil {
		return err
	}
	now := time.Now().UTC()
	l.UsedAt = &now
	return nil
}

// ConsentLinkRepository defines the interface for consent link persistence
type ConsentLinkRepository interface {
	FindByToken(token string) (*ConsentLink, error)
	Save(link *ConsentLink) error
	Update(link *ConsentLink) error
	Delete(token string) error
}
