# Beaesthetic Identity Service

Issues short-lived internal **PASETO v4.public** access tokens signed with Ed25519. It owns users, organizations, memberships, roles and permissions. APISIX (or another gateway) sends a Firebase ID token to its OAuth token-exchange endpoint; the service verifies it, resolves the local membership and mints a service-audience token.

Firebase verification checks the JWT signature against Firebase's JWKS and validates the expected Firebase project audience and issuer. Only the resulting local membership decides roles and permissions.

## Token claims

`iss`, `sub`, `aud`, `iat`, `exp`, `jti`, `kid`, `identity_type` (`human`, `service`, or `integration`), `permissions`, and `auth_method` are signed claims. Services must verify the signature, issuer, expiry, and their own expected audience before authorizing a request.

## Endpoints

All `/v1/internal/*` endpoints require `X-Internal-API-Key`.

- `POST /v1/internal/tokens` issues a token from a trusted, normalized identity.
- `POST /v1/internal/tokens/verify` verifies a token for a requested audience; this is useful for diagnostics, but services should normally verify locally.
- `GET /v1/internal/keys` returns the active and retained public keys for offline verification.
- `GET /health` is unauthenticated for Kubernetes probes.
- `POST /oauth/token` accepts `grant_type=urn:ietf:params:oauth:grant-type:token-exchange`, `subject_token_type=urn:ietf:params:oauth:token-type:jwt`, `authenticator=firebase`, `authorizer=membership`, a Firebase `subject_token`, `audience`, and `organization_id`.

`Authenticator` and `Authorizer` are independent registries: one validates and normalizes an external identity; the other maps it to local authorization. OIDC, mTLS, partner credentials, and authorization policies can be added without modifying PASETO issuance.

The migration in `migrations/000001_identity_membership.up.sql` defines users, external identities, organizations, memberships, roles, permissions, and their associations. Permission computation is server-side: clients can never provide roles or permissions to `/oauth/token`.

Example issue request:

```json
{
  "subject": "user_123",
  "identity_type": "human",
  "audience": "appointment",
  "permissions": ["appointments:read"],
  "auth_method": "provider:xyz"
}
```

## Key rotation

Set `ENV_TOKEN_ACTIVE_KEY_ID` and its 64-byte raw Ed25519 private key encoded with base64url (no padding). Keep old public keys in `ENV_TOKEN_VERIFY_KEYS_JSON`, e.g. `{"2026-06":"<base64url-public-key>"}`. Tokens carry the signing `kid`; verifiers can retain all advertised keys until all tokens signed by the old key expire.

The private key is never exposed. Store it in a Kubernetes Secret; do not place it in Helm values or source control.
