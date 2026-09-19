package firebase

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/application"
)

const firebaseJWKSURL = "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"

type FirebaseVerifier struct {
	projectID string
	client    *http.Client
}

func NewFirebaseVerifier(projectID string, client *http.Client) *FirebaseVerifier {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &FirebaseVerifier{projectID: projectID, client: client}
}

func (v *FirebaseVerifier) Name() string { return "firebase" }

func (v *FirebaseVerifier) Authenticate(ctx context.Context, subjectToken string) (application.ExternalIdentity, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithAudience(v.projectID), jwt.WithIssuer("https://securetoken.google.com/"+v.projectID))
	claims := jwt.MapClaims{}
	parsed, err := parser.ParseWithClaims(subjectToken, claims, func(token *jwt.Token) (any, error) {
		keyID, _ := token.Header["kid"].(string)
		return v.publicKey(ctx, keyID)
	})
	if err != nil || !parsed.Valid {
		return application.ExternalIdentity{}, fmt.Errorf("invalid Firebase ID token")
	}
	subject, _ := claims.GetSubject()
	if subject == "" {
		return application.ExternalIdentity{}, fmt.Errorf("Firebase token has no subject")
	}
	email, _ := claims["email"].(string)
	return application.ExternalIdentity{Provider: v.Name(), Subject: subject, Email: email}, nil
}

func (v *FirebaseVerifier) publicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	if keyID == "" {
		return nil, fmt.Errorf("Firebase token has no key ID")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, firebaseJWKSURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := v.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Firebase key endpoint returned %s", response.Status)
	}
	var document struct {
		Keys []struct {
			KeyID    string `json:"kid"`
			Modulus  string `json:"n"`
			Exponent string `json:"e"`
			Type     string `json:"kty"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return nil, err
	}
	for _, key := range document.Keys {
		if key.KeyID != keyID || key.Type != "RSA" {
			continue
		}
		n, err := base64.RawURLEncoding.DecodeString(key.Modulus)
		if err != nil {
			return nil, err
		}
		e, err := base64.RawURLEncoding.DecodeString(key.Exponent)
		if err != nil {
			return nil, err
		}
		return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}, nil
	}
	return nil, fmt.Errorf("Firebase key %q not found", keyID)
}
