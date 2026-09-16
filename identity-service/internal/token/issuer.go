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
	publicKeys  map[string]paseto.V4AsymmetricPublicKey
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
	publicKeys := make(map[string]paseto.V4AsymmetricPublicKey, len(cfg.VerifyKeys)+1)
	publicB64 := make(map[string]string, len(cfg.VerifyKeys)+1)
	activePublic := privateKey.Public()
	publicKeys[cfg.ActiveKeyID] = activePublic
	publicB64[cfg.ActiveKeyID] = base64.RawURLEncoding.EncodeToString(activePublic.ExportBytes())
	for keyID, encoded := range cfg.VerifyKeys {
		if keyID == "" || keyID == cfg.ActiveKeyID {
			continue
		}
		publicBytes, err := base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("decode public key %q: %w", keyID, err)
		}
		if len(publicBytes) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("public key %q must be %d bytes", keyID, ed25519.PublicKeySize)
		}
		publicKey, err := paseto.NewV4AsymmetricPublicKeyFromBytes(publicBytes)
		if err != nil {
			return nil, fmt.Errorf("create public key %q: %w", keyID, err)
		}
		publicKeys[keyID] = publicKey
		publicB64[keyID] = encoded
	}
	clock := cfg.Clock
	if clock == nil {
		clock = time.Now
	}
	return &Issuer{cfg.Issuer, cfg.ActiveKeyID, privateKey, publicKeys, publicB64, cfg.TTL, clock}, nil
}

func (i *Issuer) Issue(request IssueRequest) (string, Claims, error) {
	if request.Subject == "" || request.Audience == "" || request.AuthMethod == "" || !request.IdentityType.Valid() {
		return "", Claims{}, fmt.Errorf("subject, audience, auth method, and a valid identity type are required")
	}
	now := i.clock().UTC()
	tokenID, err := randomID()
	if err != nil {
		return "", Claims{}, err
	}
	claims := Claims{
		Issuer: i.issuer, Subject: request.Subject, Audience: request.Audience,
		IssuedAt: now, ExpiresAt: now.Add(i.ttl), TokenID: tokenID, KeyID: i.activeKeyID,
		IdentityType: request.IdentityType, Permissions: normalizedPermissions(request.Permissions), AuthMethod: request.AuthMethod,
		OrganizationID: request.OrganizationID, MembershipID: request.MembershipID, Roles: normalizedPermissions(request.Roles),
	}
	t := paseto.NewToken()
	t.SetIssuer(claims.Issuer)
	t.SetSubject(claims.Subject)
	t.SetAudience(claims.Audience)
	t.SetIssuedAt(claims.IssuedAt)
	t.SetExpiration(claims.ExpiresAt)
	t.SetJti(claims.TokenID)
	t.Set("kid", claims.KeyID)
	t.Set("identity_type", string(claims.IdentityType))
	t.Set("organization_id", claims.OrganizationID)
	t.Set("membership_id", claims.MembershipID)
	t.Set("roles", claims.Roles)
	t.Set("permissions", claims.Permissions)
	t.Set("auth_method", claims.AuthMethod)
	return t.V4Sign(i.privateKey, nil), claims, nil
}

func (i *Issuer) Verify(signed, audience string) (Claims, error) {
	if signed == "" || audience == "" {
		return Claims{}, fmt.Errorf("token and expected audience are required")
	}
	var lastErr error
	for keyID, publicKey := range i.publicKeys {
		token, err := paseto.NewParser().ParseV4Public(publicKey, signed, nil)
		if err != nil {
			lastErr = err
			continue
		}
		claims, err := claimsFromToken(token)
		if err != nil {
			return Claims{}, err
		}
		if claims.KeyID != keyID || claims.Issuer != i.issuer || claims.Audience != audience || claims.ExpiresAt.Before(i.clock().UTC()) || !claims.IdentityType.Valid() {
			return Claims{}, fmt.Errorf("token claims are invalid")
		}
		return claims, nil
	}
	return Claims{}, fmt.Errorf("verify PASETO: %w", lastErr)
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

func claimsFromToken(token *paseto.Token) (Claims, error) {
	var claims Claims
	if err := token.Get("kid", &claims.KeyID); err != nil {
		return Claims{}, err
	}
	if err := token.Get("identity_type", &claims.IdentityType); err != nil {
		return Claims{}, err
	}
	if err := token.Get("organization_id", &claims.OrganizationID); err != nil {
		return Claims{}, err
	}
	if err := token.Get("membership_id", &claims.MembershipID); err != nil {
		return Claims{}, err
	}
	if err := token.Get("roles", &claims.Roles); err != nil {
		return Claims{}, err
	}
	if err := token.Get("permissions", &claims.Permissions); err != nil {
		return Claims{}, err
	}
	if err := token.Get("auth_method", &claims.AuthMethod); err != nil {
		return Claims{}, err
	}
	var err error
	if claims.Issuer, err = token.GetIssuer(); err != nil {
		return Claims{}, err
	}
	if claims.Subject, err = token.GetSubject(); err != nil {
		return Claims{}, err
	}
	if claims.Audience, err = token.GetAudience(); err != nil {
		return Claims{}, err
	}
	if claims.IssuedAt, err = token.GetIssuedAt(); err != nil {
		return Claims{}, err
	}
	if claims.ExpiresAt, err = token.GetExpiration(); err != nil {
		return Claims{}, err
	}
	if claims.TokenID, err = token.GetJti(); err != nil {
		return Claims{}, err
	}
	return claims, nil
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
