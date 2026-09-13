package di

import (
	"sync"
	"testing"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/config"
	"github.com/stretchr/testify/require"
)

func TestRiverReminderConfigDefaultsSoftStopTimeout(t *testing.T) {
	container := &DiContainer{Config: config.Config{}, deps: &sync.Map{}}

	got := container.GetRiverReminderConfig()

	require.Equal(t, 8*time.Second, got.SoftStopTimeout)
}

func TestRiverReminderConfigUsesConfiguredSoftStopTimeout(t *testing.T) {
	container := &DiContainer{
		Config: config.Config{River: config.RiverConfig{SoftStopTimeout: 5 * time.Second}},
		deps:   &sync.Map{},
	}

	got := container.GetRiverReminderConfig()

	require.Equal(t, 5*time.Second, got.SoftStopTimeout)
}
