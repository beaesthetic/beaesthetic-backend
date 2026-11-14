# Shared Go core

Module: `github.com/petretiandrea/beaesthetic-backend/core` (matching the existing Go module namespace).
Reusable backend behavior belongs here; generated API contracts remain in `core-contracts` and business rules stay in their services.

## Identity authentication and authorization

`auth` consumes the `v4.public` PASETO access tokens issued by `identity-service/internal/domain/token`. It needs only public keys, never Identity's signing key. Every service uses `auth.Config`: `issuer` and `public_keys`, where `public_keys` is a JSON `kid -> public_key_b64` map of raw URL-base64 Ed25519 keys. Services may instead set `public_keys_url` and `internal_api_key` to retrieve keys from Identity's authenticated internal endpoint at startup; `public_keys` takes precedence. Recreate the verifier when rotating keys; retain old keys until their tokens expire. Unknown keys fail closed.

The verifier checks the signature before returning any data, verifies issuer and matching signed `kid`, requires subject, token ID, authentication method, supported identity type and scope, and validates `iat`/`exp` against one clock reading. Expiration is exclusive; future issuance is rejected without clock tolerance. Identity does not emit `nbf`; if supplied, it is also enforced. User scope cannot carry organization, membership or roles; membership scope requires both organization and membership IDs.

```go
verifier, err := auth.NewVerifier(auth.Config{
    Issuer:         "identity", // must match Identity configuration
    PublicKeysJSON: `{"key-1":"` + publicKeyB64 + `"}`,
})
if err != nil {
    return err // fail service startup on invalid configuration
}

policy := auth.Policy{
    Permissions: []string{"appointments.read"},
    Scope: auth.ScopeMembership,
    OrganizationID: organizationID, // expected tenant from trusted route/configuration
}
protected := verifier.Authenticate(auth.Require(policy)(handler))
mux.Handle("GET /appointments", protected)
```

Import `auth` from `github.com/petretiandrea/beaesthetic-backend/core/auth`. Sibling modules can use `require .../core v0.0.0` and `replace .../core => ../core`, consistent with the repository's existing local modules. Include `core` in build contexts when a service adopts this dependency.

The middleware uses standard `net/http`, so it can also wrap a Gin engine at the HTTP server boundary (`Handler: verifier.Authenticate(auth.Require(policy)(router))`). Use separate wrappers for routes with different policies. Apply authentication before authorization. An empty policy requires authentication only. All listed permissions must match exactly: roles, `*` and permission prefixes grant nothing implicitly. Configure a membership scope and expected organization for tenant-bound endpoints, and constrain database queries to that verified organization. For dynamic tenant routes, resolve the expected organization from the route and compare it to the verified claims before querying data.

Only one `Authorization: Bearer <token>` header is accepted (scheme is case insensitive). Missing or invalid authentication returns 401 with a Bearer challenge; failed authorization returns 403. Rejected requests never reach the protected handler. `auth.FromContext(r.Context())` returns a defensive copy of verified claims for application code. Tokens are not accepted from query strings or cookies.

Verification is offline: revocation and updated grants take effect when tokens expire or the trusted key set changes. No audience is enforced because the current Identity contract does not issue `aud`. Route policies in appointment and customer are tracked separately in SBR-83 and SBR-84.

## Checks

```sh
cd core
go test ./...
go vet ./...
```

`auth/testdata/identity-token.json` was generated using the actual Identity `NewIssuer`/`Issue` implementation at commit `644b9118017e35f575de5b27fddcaae89e1f5b0f`, with a deterministic, test-only Ed25519 seed, issuer `identity`, key ID `fixture`, and a fixed clock of `2026-11-03T12:00:00Z` (one hour TTL). It contains no production credentials. The compatibility test verifies it with a fixed clock.
