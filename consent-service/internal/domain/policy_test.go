package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPolicy(t *testing.T) {
	t.Run("creates policy with valid inputs", func(t *testing.T) {
		policy, err := NewPolicy("privacy-policy", "Privacy Policy", "Our privacy policy")

		require.NoError(t, err)
		assert.Equal(t, "privacy-policy", policy.Slug)
		assert.Equal(t, "Privacy Policy", policy.Name)
		assert.Equal(t, "Our privacy policy", policy.Description)
		assert.Empty(t, policy.Versions)
		assert.False(t, policy.CreatedAt.IsZero())
		assert.False(t, policy.UpdatedAt.IsZero())
	})

	t.Run("returns error for empty slug", func(t *testing.T) {
		policy, err := NewPolicy("", "Privacy Policy", "Description")

		assert.Nil(t, policy)
		assert.ErrorIs(t, err, ErrInvalidPolicySlug)
	})
}

func TestPolicy_AddVersion(t *testing.T) {
	t.Run("adds first version", func(t *testing.T) {
		policy, _ := NewPolicy("privacy-policy", "Privacy Policy", "Description")

		version := PolicyVersion{
			Version:         "1.0.0",
			ContentHTML:     "<h1>Test</h1>",
			ContentMarkdown: "# Test",
			PDFURL:          "https://cdn.example.com/policy.pdf",
			IsActive:        true,
		}

		err := policy.AddVersion(version)

		require.NoError(t, err)
		assert.Len(t, policy.Versions, 1)
		assert.Equal(t, "1.0.0", policy.Versions[0].Version)
		assert.True(t, policy.Versions[0].IsActive)
	})

	t.Run("deactivates previous version when adding active version", func(t *testing.T) {
		policy, _ := NewPolicy("privacy-policy", "Privacy Policy", "Description")

		policy.AddVersion(PolicyVersion{Version: "1.0.0", IsActive: true})
		policy.AddVersion(PolicyVersion{Version: "1.1.0", IsActive: true})

		assert.Len(t, policy.Versions, 2)
		assert.False(t, policy.Versions[0].IsActive)
		assert.True(t, policy.Versions[1].IsActive)
	})

	t.Run("returns error for duplicate version", func(t *testing.T) {
		policy, _ := NewPolicy("privacy-policy", "Privacy Policy", "Description")

		policy.AddVersion(PolicyVersion{Version: "1.0.0", IsActive: true})
		err := policy.AddVersion(PolicyVersion{Version: "1.0.0", IsActive: false})

		assert.ErrorIs(t, err, ErrVersionAlreadyExists)
	})

	t.Run("returns error for empty version", func(t *testing.T) {
		policy, _ := NewPolicy("privacy-policy", "Privacy Policy", "Description")

		err := policy.AddVersion(PolicyVersion{Version: "", IsActive: true})

		assert.ErrorIs(t, err, ErrInvalidPolicyVersion)
	})
}

func TestPolicy_GetActiveVersion(t *testing.T) {
	t.Run("returns active version", func(t *testing.T) {
		policy, _ := NewPolicy("privacy-policy", "Privacy Policy", "Description")
		policy.AddVersion(PolicyVersion{Version: "1.0.0", IsActive: false})
		policy.AddVersion(PolicyVersion{Version: "1.1.0", IsActive: true})

		activeVersion, err := policy.GetActiveVersion()

		require.NoError(t, err)
		assert.Equal(t, "1.1.0", activeVersion.Version)
	})

	t.Run("returns error when no active version", func(t *testing.T) {
		policy, _ := NewPolicy("privacy-policy", "Privacy Policy", "Description")
		policy.AddVersion(PolicyVersion{Version: "1.0.0", IsActive: false})

		activeVersion, err := policy.GetActiveVersion()

		assert.Nil(t, activeVersion)
		assert.ErrorIs(t, err, ErrNoActiveVersion)
	})
}

func TestPolicy_GetVersion(t *testing.T) {
	t.Run("returns specific version", func(t *testing.T) {
		policy, _ := NewPolicy("privacy-policy", "Privacy Policy", "Description")
		policy.AddVersion(PolicyVersion{Version: "1.0.0", ContentHTML: "v1"})
		policy.AddVersion(PolicyVersion{Version: "1.1.0", ContentHTML: "v2"})

		version, err := policy.GetVersion("1.0.0")

		require.NoError(t, err)
		assert.Equal(t, "v1", version.ContentHTML)
	})

	t.Run("returns error for non-existent version", func(t *testing.T) {
		policy, _ := NewPolicy("privacy-policy", "Privacy Policy", "Description")
		policy.AddVersion(PolicyVersion{Version: "1.0.0"})

		version, err := policy.GetVersion("2.0.0")

		assert.Nil(t, version)
		assert.ErrorIs(t, err, ErrInvalidPolicyVersion)
	})
}
