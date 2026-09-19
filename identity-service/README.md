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
- `GET /v1/internal/organizations/{organization_id}/users/{user_id}/authorization` returns an active membership's effective roles and permissions.
- `GET /health` is unauthenticated for Kubernetes probes.
- `POST /oauth/token` accepts JSON matching `ExchangeTokenRequest`: `grantType=urn:ietf:params:oauth:grant-type:token-exchange`, `subjectAssertion.tokenType=urn:ietf:params:oauth:token-type:jwt`, `subjectAssertion.authenticator=firebase`, `authorizer=membership`, a Firebase `subjectAssertion.token`, `audience`, and `organizationId`.

gRPC listens on `ENV_GRPC_ADDR` (default `:9090`) and implements `TokenService` and `MembershipService` from `core-contracts/identity`. `ExchangeToken` is public; every other RPC requires gRPC metadata `x-internal-api-key`.

`Authenticator` and `Authorizer` are independent registries: one validates and normalizes an external identity; the other maps it to local authorization. OIDC, mTLS, partner credentials, and authorization policies can be added without modifying PASETO issuance.

The migration in `migrations/000001_identity_membership.up.sql` defines users, external identities, organizations, memberships, roles, permissions, and their associations. Permission computation is server-side: clients can never provide roles or permissions to `/oauth/token`.

Example issue request:

```json
{
  "subject": "user_123",
  "identityType": "IDENTITY_TYPE_HUMAN",
  "audience": "appointment",
  "permissions": ["appointments:read"],
  "authMethod": "provider:xyz"
}
```

## Key rotation

Set `ENV_TOKEN_ACTIVE_KEY_ID` and its 64-byte raw Ed25519 private key encoded with base64url (no padding). Keep old public keys in `ENV_TOKEN_VERIFY_KEYS_JSON`, e.g. `{"2026-06":"<base64url-public-key>"}`. Tokens carry the signing `kid`; verifiers can retain all advertised keys until all tokens signed by the old key expire.

The private key is never exposed. Store it in a Kubernetes Secret; do not place it in Helm values or source control.

## Developer commands

Uses Mage, matching `appointment`:

```bash
mage lint   # go fmt ./... + go vet ./...
mage test
mage check
mage build
```

## Authorization seed

`seeds/roles.yaml` is the global role and permission catalog. Memberships receive roles through `membership_roles`; permissions are derived only through `role_permissions`.
