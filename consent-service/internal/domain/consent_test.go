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
			"tenant-1",
			"subject-123",
			"privacy-policy",
			"1.0.0",
			AcceptanceMethodDirect,
			nil, nil, nil,
		)

		require.NoError(t, err)
		assert.Equal(t, "consent-id", consent.ID)
		assert.Equal(t, "tenant-1", consent.TenantID)
		assert.Equal(t, "subject-123", consent.Subject)
		assert.Equal(t, "privacy-policy", consent.PolicySlug)
		assert.Equal(t, "1.0.0", consent.PolicyVersion)
		assert.Equal(t, AcceptanceMethodDirect, consent.AcceptanceMethod)
		assert.Nil(t, consent.LinkToken)
		assert.Nil(t, consent.IpAddress)
		assert.Nil(t, consent.UserAgent)
		assert.Nil(t, consent.RevokedAt)
		assert.Nil(t, consent.RevokedBy)
		assert.False(t, consent.AcceptedAt.IsZero())
	})

	t.Run("creates consent with link token", func(t *testing.T) {
		token := "link-token-123"
		consent, err := NewConsent(
			"consent-id",
			"tenant-1",
			"subject-123",
			"privacy-policy",
			"1.0.0",
			AcceptanceMethodLink,
			&token, nil, nil,
		)

		require.NoError(t, err)
		assert.Equal(t, AcceptanceMethodLink, consent.AcceptanceMethod)
		assert.NotNil(t, consent.LinkToken)
		assert.Equal(t, token, *consent.LinkToken)
	})

	t.Run("creates consent with ip and user agent", func(t *testing.T) {
		ip := "192.168.1.1"
		ua := "Mozilla/5.0"
		consent, err := NewConsent(
			"consent-id",
			"tenant-1",
			"subject-123",
			"privacy-policy",
			"1.0.0",
			AcceptanceMethodLink,
			nil, &ip, &ua,
		)

		require.NoError(t, err)
		assert.NotNil(t, consent.IpAddress)
		assert.Equal(t, ip, *consent.IpAddress)
		assert.NotNil(t, consent.UserAgent)
		assert.Equal(t, ua, *consent.UserAgent)
	})

	t.Run("returns error for empty tenant id", func(t *testing.T) {
		consent, err := NewConsent("id", "", "subject", "policy", "1.0.0", AcceptanceMethodDirect, nil, nil, nil)

		assert.Nil(t, consent)
		assert.ErrorIs(t, err, ErrInvalidTenantID)
	})

	t.Run("returns error for empty subject", func(t *testing.T) {
		consent, err := NewConsent("id", "tenant-1", "", "policy", "1.0.0", AcceptanceMethodDirect, nil, nil, nil)

		assert.Nil(t, consent)
		assert.ErrorIs(t, err, ErrInvalidSubject)
	})

	t.Run("returns error for empty policy slug", func(t *testing.T) {
		consent, err := NewConsent("id", "tenant-1", "subject", "", "1.0.0", AcceptanceMethodDirect, nil, nil, nil)

		assert.Nil(t, consent)
		assert.ErrorIs(t, err, ErrInvalidPolicySlug)
	})

	t.Run("returns error for empty policy version", func(t *testing.T) {
		consent, err := NewConsent("id", "tenant-1", "subject", "policy", "", AcceptanceMethodDirect, nil, nil, nil)

		assert.Nil(t, consent)
		assert.ErrorIs(t, err, ErrInvalidPolicyVersion)
	})
}

func TestConsent_Revoke(t *testing.T) {
	t.Run("revokes active consent", func(t *testing.T) {
		consent, _ := NewConsent("id", "tenant-1", "subject", "policy", "1.0.0", AcceptanceMethodDirect, nil, nil, nil)

		err := consent.Revoke("operator-123")

		require.NoError(t, err)
		assert.NotNil(t, consent.RevokedAt)
		assert.NotNil(t, consent.RevokedBy)
		assert.Equal(t, "operator-123", *consent.RevokedBy)
		assert.False(t, consent.IsActive())
	})

	t.Run("returns error when already revoked", func(t *testing.T) {
		consent, _ := NewConsent("id", "tenant-1", "subject", "policy", "1.0.0", AcceptanceMethodDirect, nil, nil, nil)
		consent.Revoke("operator-1")

		err := consent.Revoke("operator-2")

		assert.ErrorIs(t, err, ErrConsentAlreadyRevoked)
	})
}

func TestConsent_IsActive(t *testing.T) {
	t.Run("returns true for non-revoked consent", func(t *testing.T) {
		consent, _ := NewConsent("id", "tenant-1", "subject", "policy", "1.0.0", AcceptanceMethodDirect, nil, nil, nil)

		assert.True(t, consent.IsActive())
	})

	t.Run("returns false for revoked consent", func(t *testing.T) {
		consent, _ := NewConsent("id", "tenant-1", "subject", "policy", "1.0.0", AcceptanceMethodDirect, nil, nil, nil)
		consent.Revoke("operator")

		assert.False(t, consent.IsActive())
	})
}
