package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsentLink(t *testing.T) {
	t.Run("creates link with valid inputs", func(t *testing.T) {
		link, err := NewConsentLink(
			"token-123",
			"subject-123",
			[]string{"privacy-policy", "marketing-consent"},
			48,
			"operator-123",
		)

		require.NoError(t, err)
		assert.Equal(t, "token-123", link.Token)
		assert.Equal(t, "subject-123", link.Subject)
		assert.Len(t, link.Policies, 2)
		assert.Equal(t, "operator-123", link.CreatedBy)
		assert.Nil(t, link.UsedAt)
		assert.False(t, link.CreatedAt.IsZero())
		assert.True(t, link.ExpiresAt.After(link.CreatedAt))
	})

	t.Run("sets correct expiry time", func(t *testing.T) {
		link, _ := NewConsentLink("token", "subject", []string{"policy"}, 24, "operator")

		expectedExpiry := link.CreatedAt.Add(24 * time.Hour)
		assert.Equal(t, expectedExpiry.Unix(), link.ExpiresAt.Unix())
	})

	t.Run("returns error for empty token", func(t *testing.T) {
		link, err := NewConsentLink("", "subject", []string{"policy"}, 48, "operator")

		assert.Nil(t, link)
		assert.ErrorIs(t, err, ErrInvalidLinkToken)
	})

	t.Run("returns error for empty subject", func(t *testing.T) {
		link, err := NewConsentLink("token", "", []string{"policy"}, 48, "operator")

		assert.Nil(t, link)
		assert.ErrorIs(t, err, ErrInvalidSubject)
	})

	t.Run("returns error for empty policies", func(t *testing.T) {
		link, err := NewConsentLink("token", "subject", []string{}, 48, "operator")

		assert.Nil(t, link)
		assert.ErrorIs(t, err, ErrInvalidPolicySlug)
	})
}

func TestConsentLink_IsValid(t *testing.T) {
	t.Run("returns nil for valid link", func(t *testing.T) {
		link, _ := NewConsentLink("token", "subject", []string{"policy"}, 48, "operator")

		err := link.IsValid()

		assert.NoError(t, err)
	})

	t.Run("returns error for expired link", func(t *testing.T) {
		link, _ := NewConsentLink("token", "subject", []string{"policy"}, 1, "operator")
		link.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour)

		err := link.IsValid()

		assert.ErrorIs(t, err, ErrLinkExpired)
	})

	t.Run("returns error for used link", func(t *testing.T) {
		link, _ := NewConsentLink("token", "subject", []string{"policy"}, 48, "operator")
		now := time.Now().UTC()
		link.UsedAt = &now

		err := link.IsValid()

		assert.ErrorIs(t, err, ErrLinkAlreadyUsed)
	})
}

func TestConsentLink_MarkAsUsed(t *testing.T) {
	t.Run("marks valid link as used", func(t *testing.T) {
		link, _ := NewConsentLink("token", "subject", []string{"policy"}, 48, "operator")

		err := link.MarkAsUsed()

		require.NoError(t, err)
		assert.NotNil(t, link.UsedAt)
	})

	t.Run("returns error for expired link", func(t *testing.T) {
		link, _ := NewConsentLink("token", "subject", []string{"policy"}, 1, "operator")
		link.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour)

		err := link.MarkAsUsed()

		assert.ErrorIs(t, err, ErrLinkExpired)
	})

	t.Run("returns error for already used link", func(t *testing.T) {
		link, _ := NewConsentLink("token", "subject", []string{"policy"}, 48, "operator")
		link.MarkAsUsed()

		err := link.MarkAsUsed()

		assert.ErrorIs(t, err, ErrLinkAlreadyUsed)
	})
}
