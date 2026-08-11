package postgres

import (
	"strings"
	"testing"
)

func TestValidateLegacyAgendaBackfillRejectsInvalidCustomerIDs(t *testing.T) {
	err := validateLegacyAgendaBackfill(LegacyAgendaBackfill{SkippedInvalidCustomers: 2})
	if err == nil {
		t.Fatal("expected invalid customer IDs to block the backfill")
	}
}

func TestValidateLegacyAgendaBackfillAcceptsMigratableData(t *testing.T) {
	if err := validateLegacyAgendaBackfill(LegacyAgendaBackfill{}); err != nil {
		t.Fatalf("validateLegacyAgendaBackfill() error = %v", err)
	}
}

func TestCalendarEventBaseBackfillPreservesLegacyDisplayAndCancellation(t *testing.T) {
	for _, fragment := range []string{
		"display_title = coalesce(display_title, nullif(title, ''))",
		"display_description = coalesce(display_description, nullif(description, ''))",
		"cancel_reason IS NOT NULL THEN coalesce(canceled_at, updated_at)",
	} {
		if !strings.Contains(normalizeCalendarEventBaseSQL, fragment) {
			t.Fatalf("calendar event base backfill does not contain %q", fragment)
		}
	}
}
