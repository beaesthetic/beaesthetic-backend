package postgres

import (
	"testing"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

func TestPersistedLegacyEventTypeUsesV2ManualValue(t *testing.T) {
	if got := persistedLegacyEventType(domain.EventTypeGeneric); got != "manual" {
		t.Fatalf("persistedLegacyEventType(generic) = %q, want manual", got)
	}
	if got := persistedLegacyEventType(domain.EventTypeAppointment); got != "appointment" {
		t.Fatalf("persistedLegacyEventType(appointment) = %q, want appointment", got)
	}
}
