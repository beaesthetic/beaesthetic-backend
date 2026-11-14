# Customer service

The HTTP API verifies Identity `v4.public` PASETO tokens through `core/auth`.
All customer, fidelity card and wallet endpoints (including trailing slash
aliases) require a verified membership token:

| Operations | Permission |
| --- | --- |
| All GET endpoints and POST `/customers/search:byPhone` | `customer.read` |
| All other POST, PUT and DELETE endpoints | `customer.write` |

Permissions match exactly; owner/admin roles and wildcard strings do not grant
access by themselves. Missing/invalid tokens return 401 with a Bearer challenge;
insufficient permissions or user scope return 403 before request binding, cache
or database access. `/health` remains public.

Required application configuration (environment or `--env-file`):

```dotenv
ENV_AUTH_ISSUER=identity
ENV_AUTH_PUBLIC__KEYS='{"<active-kid>":"<public-key-base64url>"}'
```

`ENV_AUTH_PUBLIC__KEYS` is a JSON `kid -> public_key_b64` map. Include the
active Identity public key and retained old keys until their tokens expire.
Invalid or missing configuration prevents API startup; migrations do not require
authentication configuration. No private signing key is needed. Helm deployments
configure the public-key map in `ENV_AUTH_PUBLIC__KEYS`.

Customer currently serves a single organization, and its database/cache do not
partition data by tenant. The service neither requires nor uses an organization
header, and applies no organization-specific authorization constraint. The core
verifier still validates membership token structure, including `organization_id`,
and exposes the verified claims in request context.

Internal callers, including appointment and notification, must send
`Authorization: Bearer <token>` with the appropriate customer permission.
Tenant isolation will need database/cache partitioning and organization policies
before this service can support multiple organizations.

Run checks with `cd customer && go test ./... && go vet ./...`.
Build the container from the repository root:
`docker build -f customer/Dockerfile .` (the build includes the sibling core module).
