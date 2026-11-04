# Customer service

The HTTP API verifies Identity `v4.public` PASETO tokens through `core/auth`.
All customer, fidelity card and wallet endpoints (including trailing slash
aliases) require a membership token and an `X-Organization-ID` request header.
The header must match the token’s `organization_id`:

| Operations | Permission |
| --- | --- |
| All GET endpoints and POST `/customers/search:byPhone` | `customer.read` |
| All other POST, PUT and DELETE endpoints | `customer.write` |

Permissions match exactly; owner/admin roles and wildcard strings do not grant
access by themselves. Missing/invalid tokens return 401 with a Bearer challenge;
insufficient permissions, user scope or a mismatched organization return 403 before
request binding, cache or database access. Missing, empty or duplicate organization
headers return 400 after authentication. `/health` remains public.

Required application configuration (environment or `--env-file`):

```dotenv
ENV_AUTH_ISSUER=identity
ENV_AUTH_PUBLIC__KEYS={"key-1":"<raw URL-base64 Ed25519 public key>"}
```

Public keys must come from a trusted Identity configuration. Key IDs are preserved
exactly in the JSON map. Restart with the new key set for rotation, retaining old
keys until their tokens expire. Invalid/missing configuration prevents API startup;
migrations do not require authentication configuration. No private signing key is
needed. Helm deployments supply these variables through `envConfig` or
`secretKeyRef` before rolling out this version.

Organization selection is per request, with no fixed organization in service
configuration. Internal callers, including appointment and notification, must send
both `Authorization: Bearer <token>` and `X-Organization-ID`.

The header/token comparison authorizes the requested organization. The existing
repositories and cache do not partition data by organization; this change does not
add database/cache tenant isolation.

Run checks with `cd customer && go test ./... && go vet ./...`.
Build the container from the repository root:
`docker build -f customer/Dockerfile .` (the build includes the sibling core module).
