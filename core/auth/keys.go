package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const verificationKeysRequestTimeout = 5 * time.Second

// FetchPublicKeys retrieves keys from Identity's authenticated internal endpoint.
func FetchPublicKeys(url, internalAPIKey string) (map[string]string, error) {
	if strings.TrimSpace(url) == "" || strings.TrimSpace(internalAPIKey) == "" {
		return nil, fmt.Errorf("verification keys URL and internal API key are required")
	}
	client := &http.Client{Timeout: verificationKeysRequestTimeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-Internal-API-Key", internalAPIKey)
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request verification keys: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request verification keys: unexpected status %s", response.Status)
	}
	var payload struct {
		Keys []struct {
			KeyID        string `json:"kid"`
			PublicKeyB64 string `json:"publicKeyB64"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode verification keys: %w", err)
	}
	keys := make(map[string]string, len(payload.Keys))
	for _, key := range payload.Keys {
		if key.KeyID == "" || key.PublicKeyB64 == "" {
			return nil, fmt.Errorf("verification key ID and public key are required")
		}
		if _, duplicate := keys[key.KeyID]; duplicate {
			return nil, fmt.Errorf("duplicate verification key %q", key.KeyID)
		}
		keys[key.KeyID] = key.PublicKeyB64
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("verification keys response is empty")
	}
	return keys, nil
}
