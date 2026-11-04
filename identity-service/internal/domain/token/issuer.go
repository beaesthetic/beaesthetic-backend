package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
)

type Config struct {
	Issuer        string
	ActiveKeyID   string
	PrivateKeyB64 string
	VerifyKeys    map[string]string
	TTL           time.Duration
	Clock         func() time.Time
}

type Issuer struct {
	issuer      string
	activeKeyID string
	privateKey  paseto.V4AsymmetricSecretKey
	verifier    *auth.Verifier
	publicB64   map[string]string
	ttl         time.Duration
	clock       func() time.Time
}

func NewIssuer(cfg Config) (*Issuer, error) {
	if cfg.Issuer == "" || cfg.ActiveKeyID == "" || cfg.PrivateKeyB64 == "" || cfg.TTL <= 0 {
		return nil, fmt.Errorf("issuer, active key ID, private key, and positive TTL are required")
	}
	privateBytes, err := base64.RawURLEncoding.DecodeString(cfg.PrivateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("decode active private key: %w", err)
	}
	if len(privateBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("active private key must be %d bytes", ed25519.PrivateKeySize)
	}
	privateKey, err := paseto.NewV4AsymmetricSecretKeyFromBytes(privateBytes)
	if err != nil {
		return nil, fmt.Errorf("create active private key: %w", err)
	}
	publicB64 := make(map[string]string, len(cfg.VerifyKeys)+1)
	activePublic := privateKey.Public()
	publicB64[cfg.ActiveKeyID] = base64.RawURLEncoding.EncodeToString(activePublic.ExportBytes())
	for keyID, encoded := range cfg.VerifyKeys {
		if keyID == "" || keyID == cfg.ActiveKeyID {
			continue
		}
		publicB64[keyID] = encoded
	}
	clock := cfg.Clock
	if clock == nil {
		clock = time.Now
	}
	verifier, err := auth.NewVerifier(auth.Config{Issuer: cfg.Issuer, PublicKeys: publicB64, Clock: clock})
	if err != nil {
		return nil, err
	}
	return &Issuer{cfg.Issuer, cfg.ActiveKeyID, privateKey, verifier, publicB64, cfg.TTL, clock}, nil
}

func (i *Issuer) Issue(request IssueRequest) (string, auth.Claims, error) {
	if request.Subject == "" || request.AuthMethod == "" || !request.IdentityType.Valid() || !request.Scope.Valid() {
		return "", auth.Claims{}, fmt.Errorf("subject, auth method, valid identity type, and valid scope are required")
	}
	if request.Scope == auth.ScopeUser && (request.OrganizationID != "" || request.MembershipID != "" || len(request.Roles) != 0) {
		return "", auth.Claims{}, fmt.Errorf("user-scoped token cannot contain membership claims")
	}
	if request.Scope == auth.ScopeMembership && (request.OrganizationID == "" || request.MembershipID == "") {
		return "", auth.Claims{}, fmt.Errorf("membership-scoped token requires organization and membership IDs")
	}
	now := i.clock().UTC()
	tokenID, err := randomID()
	if err != nil {
		return "", auth.Claims{}, err
	}
	claims := auth.Claims{
		Issuer: i.issuer, Subject: request.Subject,
		IssuedAt: now, ExpiresAt: now.Add(i.ttl), TokenID: tokenID, KeyID: i.activeKeyID,
		IdentityType: request.IdentityType, Scope: request.Scope, Permissions: normalizedPermissions(request.Permissions), AuthMethod: request.AuthMethod,
		OrganizationID: request.OrganizationID, MembershipID: request.MembershipID, Roles: normalizedPermissions(request.Roles),
	}
	t := paseto.NewToken()
	t.SetIssuer(claims.Issuer)
	t.SetSubject(claims.Subject)
	t.SetIssuedAt(claims.IssuedAt)
	t.SetExpiration(claims.ExpiresAt)
	t.SetJti(claims.TokenID)
	t.Set("kid", claims.KeyID)
	t.Set("identity_type", string(claims.IdentityType))
	t.Set("token_scope", string(claims.Scope))
	t.Set("organization_id", claims.OrganizationID)
	t.Set("membership_id", claims.MembershipID)
	t.Set("roles", claims.Roles)
	t.Set("permissions", claims.Permissions)
	t.Set("auth_method", claims.AuthMethod)
	return t.V4Sign(i.privateKey, nil), claims, nil
}

// Verifier exposes only public-key verification for access-token middleware.
func (i *Issuer) Verifier() *auth.Verifier { return i.verifier }

func (i *Issuer) Verify(signed string) (auth.Claims, error) {
	return i.verifier.Verify(signed)
}

func (i *Issuer) PublicKeys() []PublicKey {
	keys := make([]PublicKey, 0, len(i.publicB64))
	for keyID, publicKey := range i.publicB64 {
		keys = append(keys, PublicKey{KeyID: keyID, PublicKeyB64: publicKey})
	}
	sort.Slice(keys, func(a, b int) bool { return keys[a].KeyID < keys[b].KeyID })
	return keys
}

type PublicKey struct {
	KeyID        string `json:"kid"`
	PublicKeyB64 string `json:"public_key_b64"`
}

func normalizedPermissions(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			seen[value] = struct{}{}
		}
	}
	permissions := make([]string, 0, len(seen))
	for value := range seen {
		permissions = append(permissions, value)
	}
	sort.Strings(permissions)
	return permissions
}

func randomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate token ID: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
