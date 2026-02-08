package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsent(t *testing.T) {
	t.Run("creates consent with valid inputs", func(t *testing.T) {
		consent, err := NewConsent(
			"consent-id",
			"subject-123",
			"privacy-policy",
			"1.0.0",
			AcceptanceMethodDirect,
			nil,
		)

		require.NoError(t, err)
		assert.Equal(t, "consent-id", consent.ID)
		assert.Equal(t, "subject-123", consent.Subject)
		assert.Equal(t, "privacy-policy", consent.PolicySlug)
		assert.Equal(t, "1.0.0", consent.PolicyVersion)
		assert.Equal(t, AcceptanceMethodDirect, consent.AcceptanceMethod)
		assert.Nil(t, consent.LinkToken)
		assert.Nil(t, consent.RevokedAt)
		assert.Nil(t, consent.RevokedBy)
		assert.False(t, consent.AcceptedAt.IsZero())
	})

	t.Run("creates consent with link token", func(t *testing.T) {
		token := "link-token-123"
		consent, err := NewConsent(
			"consent-id",
			"subject-123",
			"privacy-policy",
			"1.0.0",
			AcceptanceMethodLink,
			&token,
		)

		require.NoError(t, err)
		assert.Equal(t, AcceptanceMethodLink, consent.AcceptanceMethod)
		assert.NotNil(t, consent.LinkToken)
		assert.Equal(t, token, *consent.LinkToken)
	})

	t.Run("returns error for empty subject", func(t *testing.T) {
		consent, err := NewConsent("id", "", "policy", "1.0.0", AcceptanceMethodDirect, nil)

		assert.Nil(t, consent)
		assert.ErrorIs(t, err, ErrInvalidSubject)
	})

	t.Run("returns error for empty policy slug", func(t *testing.T) {
		consent, err := NewConsent("id", "subject", "", "1.0.0", AcceptanceMethodDirect, nil)

		assert.Nil(t, consent)
		assert.ErrorIs(t, err, ErrInvalidPolicySlug)
	})

	t.Run("returns error for empty policy version", func(t *testing.T) {
		consent, err := NewConsent("id", "subject", "policy", "", AcceptanceMethodDirect, nil)

		assert.Nil(t, consent)
		assert.ErrorIs(t, err, ErrInvalidPolicyVersion)
	})
}

func TestConsent_Revoke(t *testing.T) {
	t.Run("revokes active consent", func(t *testing.T) {
		consent, _ := NewConsent("id", "subject", "policy", "1.0.0", AcceptanceMethodDirect, nil)

		err := consent.Revoke("operator-123")

		require.NoError(t, err)
		assert.NotNil(t, consent.RevokedAt)
		assert.NotNil(t, consent.RevokedBy)
		assert.Equal(t, "operator-123", *consent.RevokedBy)
		assert.False(t, consent.IsActive())
	})

	t.Run("returns error when already revoked", func(t *testing.T) {
		consent, _ := NewConsent("id", "subject", "policy", "1.0.0", AcceptanceMethodDirect, nil)
		consent.Revoke("operator-1")

		err := consent.Revoke("operator-2")

		assert.ErrorIs(t, err, ErrConsentAlreadyRevoked)
	})
}

func TestConsent_IsActive(t *testing.T) {
	t.Run("returns true for non-revoked consent", func(t *testing.T) {
		consent, _ := NewConsent("id", "subject", "policy", "1.0.0", AcceptanceMethodDirect, nil)

		assert.True(t, consent.IsActive())
	})

	t.Run("returns false for revoked consent", func(t *testing.T) {
		consent, _ := NewConsent("id", "subject", "policy", "1.0.0", AcceptanceMethodDirect, nil)
		consent.Revoke("operator")

		assert.False(t, consent.IsActive())
	})
}
