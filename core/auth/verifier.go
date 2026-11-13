// Package auth verifies Identity PASETO tokens and authorizes HTTP requests.
package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
)

// ErrInvalidToken intentionally does not expose signature or claim details.
var ErrInvalidToken = errors.New("invalid access token")

// Config is the shared service configuration for Identity PASETO verification.
// PublicKeysJSON preserves key IDs exactly, including dots and underscores.
type Config struct {
	Issuer         string `koanf:"issuer"`
	PublicKeysJSON string `koanf:"public_keys"`
}

// NewVerifier builds a verifier from service configuration. PublicKeysJSON is
// a JSON object whose keys are Identity key IDs and whose values are raw,
// URL-base64 Ed25519 public keys.
func (c Config) NewVerifier() (*Verifier, error) {
	var keys map[string]string
	if err := json.Unmarshal([]byte(c.PublicKeysJSON), &keys); err != nil {
		return nil, fmt.Errorf("invalid auth public keys: %w", err)
	}
	return NewVerifier(VerifierConfig{Issuer: c.Issuer, PublicKeys: keys})
}

// VerifierConfig is the programmatic configuration for constructing a verifier.
// Keys are copied at construction; create a new Verifier when rotating the key set.
type VerifierConfig struct {
	Issuer     string
	PublicKeys map[string]string
	Clock      func() time.Time
}

type Verifier struct {
	issuer string
	keys   map[string]paseto.V4AsymmetricPublicKey
	clock  func() time.Time
}

func NewVerifier(cfg VerifierConfig) (*Verifier, error) {
	if cfg.Issuer == "" || len(cfg.PublicKeys) == 0 {
		return nil, fmt.Errorf("issuer and public keys are required")
	}
	v := &Verifier{issuer: cfg.Issuer, keys: make(map[string]paseto.V4AsymmetricPublicKey), clock: cfg.Clock}
	if v.clock == nil {
		v.clock = time.Now
	}
	for id, encoded := range cfg.PublicKeys {
		if id == "" {
			return nil, fmt.Errorf("key ID is required")
		}
		raw, err := base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("decode public key %q: %w", id, err)
		}
		key, err := paseto.NewV4AsymmetricPublicKeyFromBytes(raw)
		if err != nil {
			return nil, fmt.Errorf("public key %q: %w", id, err)
		}
		v.keys[id] = key
	}
	return v, nil
}

// Verify only accepts signed v4.public tokens. No unverified claim is returned.
func (v *Verifier) Verify(signed string) (Claims, error) {
	now := v.clock().UTC()
	for id, key := range v.keys {
		parser := paseto.NewParserWithoutExpiryCheck()
		token, err := parser.ParseV4Public(key, signed, nil)
		if err != nil {
			continue
		}
		var payload struct {
			Claims
			NotBefore *time.Time `json:"nbf"`
		}
		if err := json.Unmarshal(token.ClaimsJSON(), &payload); err != nil {
			return Claims{}, ErrInvalidToken
		}
		c := payload.Claims
		if payload.NotBefore != nil && payload.NotBefore.After(now) {
			return Claims{}, ErrInvalidToken
		}
		if c.KeyID != id {
			continue
		}
		if c.Issuer != v.issuer || c.Subject == "" || c.TokenID == "" || c.AuthMethod == "" || !c.IdentityType.Valid() || !c.Scope.Valid() || c.IssuedAt.IsZero() || c.ExpiresAt.IsZero() || c.IssuedAt.After(now) || !c.ExpiresAt.After(now) || !c.ExpiresAt.After(c.IssuedAt) {
			return Claims{}, ErrInvalidToken
		}
		if c.Scope == ScopeUser && (c.OrganizationID != "" || c.MembershipID != "" || len(c.Roles) != 0) {
			return Claims{}, ErrInvalidToken
		}
		if c.Scope == ScopeMembership && (c.OrganizationID == "" || c.MembershipID == "") {
			return Claims{}, ErrInvalidToken
		}
		// Identity always emits these fields, including empty arrays and strings.
		for _, field := range []string{"permissions", "roles"} {
			var values []string
			if err := token.Get(field, &values); err != nil {
				return Claims{}, ErrInvalidToken
			}
		}
		return c, nil
	}
	return Claims{}, ErrInvalidToken
}

// HasPermission performs exact matching: roles and wildcard strings grant nothing implicitly.
func (c Claims) HasPermission(permission string) bool {
	if permission == "" {
		return false
	}
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}
