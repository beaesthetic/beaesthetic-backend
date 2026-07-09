package server

import (
	"errors"
	"fmt"
	"time"
)

var errMissingBody = errors.New("request body is required")

func errNotFound(resource string) error {
	return fmt.Errorf("%s not found", resource)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intValue(value *int, fallback int) int {
	if value == nil || *value <= 0 {
		return fallback
	}
	return *value
}

func stringPtrIfNotEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func timePtrIfNotZero(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
