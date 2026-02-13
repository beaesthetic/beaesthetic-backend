package domain

// ConsentStatusValue represents the status of a consent for a specific policy
type ConsentStatusValue string

const (
	ConsentStatusAccepted ConsentStatusValue = "accepted"
	ConsentStatusMissing  ConsentStatusValue = "missing"
	ConsentStatusOutdated ConsentStatusValue = "outdated"
)

// PolicyConsentStatus represents the consent status for a single policy
type PolicyConsentStatus struct {
	Slug             string
	Name             string
	Description      string
	Status           ConsentStatusValue
	ActiveVersion    string
	ConsentedVersion *string // nil when status is "missing"
}

// ComputeConsentStatus determines the consent status for a policy given a consent (or nil).
// activeVersion is the currently active PolicyVersion of the policy.
// consent is the subject's active (non-revoked) consent for this policy, or nil if none exists.
func ComputeConsentStatus(policy Policy, activeVersion PolicyVersion, consent *Consent) PolicyConsentStatus {
	status := PolicyConsentStatus{
		Slug:          policy.Slug,
		Name:          policy.Name,
		Description:   policy.Description,
		ActiveVersion: activeVersion.Version,
	}

	if consent == nil {
		status.Status = ConsentStatusMissing
		return status
	}

	status.ConsentedVersion = &consent.PolicyVersion

	if consent.PolicyVersion == activeVersion.Version {
		status.Status = ConsentStatusAccepted
		return status
	}

	// Consent is on a different version than active
	if activeVersion.RequiresReAcceptance {
		status.Status = ConsentStatusOutdated
	} else {
		status.Status = ConsentStatusAccepted
	}

	return status
}
