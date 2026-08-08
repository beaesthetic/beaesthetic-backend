package main

import (
	"testing"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

func TestReminderEligibilityForManualRiverScheduling(t *testing.T) {
	tests := []struct {
		status domain.ReminderStatus
		want   bool
	}{
		{status: domain.ReminderPending, want: true},
		{status: domain.ReminderScheduled, want: true},
		{status: domain.ReminderSentRequested, want: false},
		{status: domain.ReminderSent, want: false},
		{status: domain.ReminderFailToSend, want: false},
		{status: domain.ReminderUnprocessable, want: false},
		{status: domain.ReminderDeleted, want: false},
	}

	for _, test := range tests {
		t.Run(string(test.status), func(t *testing.T) {
			if got := isReminderEligibleForScheduling(test.status); got != test.want {
				t.Fatalf("isReminderEligibleForScheduling(%q) = %v, want %v", test.status, got, test.want)
			}
		})
	}
}
