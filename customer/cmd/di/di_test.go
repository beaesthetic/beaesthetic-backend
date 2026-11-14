package di

import (
	"encoding/base64"
	"sync"
	"testing"

	"aidanwoods.dev/go-paseto"
	"github.com/petretiandrea/beaesthetic-backend/core/auth"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/config"
	"go.uber.org/zap"
)

func TestGetAuthVerifierReturnsSingleton(t *testing.T) {
	key := paseto.NewV4AsymmetricSecretKey()
	container := &DiContainer{
		Config: config.Config{Auth: auth.Config{
			Issuer:         "identity",
			PublicKeysJSON: `{"key":"` + base64.RawURLEncoding.EncodeToString(key.Public().ExportBytes()) + `"}`,
		}},
		Log:  zap.NewNop(),
		deps: &sync.Map{},
	}
	first, err := container.GetAuthVerifier()
	if err != nil {
		t.Fatal(err)
	}
	second, err := container.GetAuthVerifier()
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("auth verifier was not cached")
	}
}
