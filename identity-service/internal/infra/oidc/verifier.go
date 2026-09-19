package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/application"
)

type Verifier struct {
	name     string
	verifier *gooidc.IDTokenVerifier
}

func NewVerifier(ctx context.Context, name, issuer, audience, discoveryURL string) (*Verifier, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create OIDC discovery request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch OIDC discovery: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery returned %s", response.Status)
	}
	var metadata gooidc.ProviderConfig
	if err := json.NewDecoder(response.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("decode OIDC discovery: %w", err)
	}
	if metadata.IssuerURL != issuer {
		return nil, fmt.Errorf("OIDC issuer mismatch: expected %q, got %q", issuer, metadata.IssuerURL)
	}
	provider := metadata.NewProvider(ctx)
	return &Verifier{name: name, verifier: provider.Verifier(&gooidc.Config{ClientID: audience})}, nil
}

func (v *Verifier) Name() string { return v.name }

func (v *Verifier) Authenticate(ctx context.Context, subjectToken string) (application.ExternalIdentity, error) {
	idToken, err := v.verifier.Verify(ctx, subjectToken)
	if err != nil {
		return application.ExternalIdentity{}, fmt.Errorf("verify OIDC token: %w", err)
	}
	var claims struct {
		Subject string `json:"sub"`
		Email   string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return application.ExternalIdentity{}, fmt.Errorf("decode OIDC claims: %w", err)
	}
	if claims.Subject == "" {
		return application.ExternalIdentity{}, fmt.Errorf("OIDC token has no subject")
	}
	return application.ExternalIdentity{Provider: v.name, Subject: claims.Subject, Email: claims.Email}, nil
}
