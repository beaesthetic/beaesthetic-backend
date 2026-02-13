package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeConsentStatus(t *testing.T) {
	policy := Policy{
		Slug:        "privacy-policy",
		Name:        "Privacy Policy",
		Description: "Our privacy policy",
	}

	t.Run("returns missing when no consent exists", func(t *testing.T) {
		activeVersion := PolicyVersion{Version: "1.0.0", IsActive: true}

		status := ComputeConsentStatus(policy, activeVersion, nil)

		assert.Equal(t, ConsentStatusMissing, status.Status)
		assert.Equal(t, "privacy-policy", status.Slug)
		assert.Equal(t, "Privacy Policy", status.Name)
		assert.Equal(t, "1.0.0", status.ActiveVersion)
		assert.Nil(t, status.ConsentedVersion)
	})

	t.Run("returns accepted when consent matches active version", func(t *testing.T) {
		activeVersion := PolicyVersion{Version: "1.0.0", IsActive: true}
		consent := &Consent{PolicyVersion: "1.0.0"}

		status := ComputeConsentStatus(policy, activeVersion, consent)

		assert.Equal(t, ConsentStatusAccepted, status.Status)
		assert.NotNil(t, status.ConsentedVersion)
		assert.Equal(t, "1.0.0", *status.ConsentedVersion)
	})

	t.Run("returns outdated when consent on old version and re-acceptance required", func(t *testing.T) {
		activeVersion := PolicyVersion{
			Version:              "2.0.0",
			IsActive:             true,
			RequiresReAcceptance: true,
		}
		consent := &Consent{PolicyVersion: "1.0.0"}

		status := ComputeConsentStatus(policy, activeVersion, consent)

		assert.Equal(t, ConsentStatusOutdated, status.Status)
		assert.Equal(t, "2.0.0", status.ActiveVersion)
		assert.NotNil(t, status.ConsentedVersion)
		assert.Equal(t, "1.0.0", *status.ConsentedVersion)
	})

	t.Run("returns accepted when consent on old version but re-acceptance not required", func(t *testing.T) {
		activeVersion := PolicyVersion{
			Version:              "1.1.0",
			IsActive:             true,
			RequiresReAcceptance: false,
		}
		consent := &Consent{PolicyVersion: "1.0.0"}

		status := ComputeConsentStatus(policy, activeVersion, consent)

		assert.Equal(t, ConsentStatusAccepted, status.Status)
		assert.Equal(t, "1.1.0", status.ActiveVersion)
		assert.NotNil(t, status.ConsentedVersion)
		assert.Equal(t, "1.0.0", *status.ConsentedVersion)
	})
}
