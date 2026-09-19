package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"
)

func TestIssueAndVerify(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	issuer, err := NewIssuer(Config{Issuer: "identity", ActiveKeyID: "key-1", PrivateKeyB64: base64.RawURLEncoding.EncodeToString(privateKey), TTL: 5 * time.Minute, Clock: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	signed, claims, err := issuer.Issue(IssueRequest{Subject: "person-1", IdentityType: IdentityHuman, Audience: "appointment", Permissions: []string{"appointments:read", " appointments:read ", "appointments:write"}, AuthMethod: "provider:xyz"})
	if err != nil {
		t.Fatal(err)
	}
	if claims.KeyID != "key-1" || len(claims.Permissions) != 2 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	verified, err := issuer.Verify(signed, "appointment")
	if err != nil {
		t.Fatal(err)
	}
	if verified.Subject != "person-1" || verified.Audience != "appointment" {
		t.Fatalf("unexpected verified claims: %+v", verified)
	}
}

func TestVerifyRejectsWrongAudienceAndTampering(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer, err := NewIssuer(Config{Issuer: "identity", ActiveKeyID: "key-1", PrivateKeyB64: base64.RawURLEncoding.EncodeToString(privateKey), TTL: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	signed, _, err := issuer.Issue(IssueRequest{Subject: "service-1", IdentityType: IdentityService, Audience: "notification", AuthMethod: "client_credentials"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := issuer.Verify(signed, "appointment"); err == nil {
		t.Fatal("wrong audience verified")
	}
	if _, err := issuer.Verify(signed+"x", "notification"); err == nil {
		t.Fatal("tampered token verified")
	}
}
