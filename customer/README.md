# Customer service

The HTTP API verifies Identity `v4.public` PASETO tokens through `core/auth`.
All customer, fidelity card and wallet endpoints (including trailing slash
aliases) require a membership token for the configured organization:

| Operations | Permission |
| --- | --- |
| All GET endpoints and POST `/customers/search:byPhone` | `customer.read` |
| All other POST, PUT and DELETE endpoints | `customer.write` |

Permissions match exactly; owner/admin roles and wildcard strings do not grant
access by themselves. Missing/invalid tokens return 401 with a Bearer challenge;
insufficient permissions, user scope or another organization return 403 before
request binding, cache or database access. `/health` remains public.

Required application configuration (environment or `--env-file`):

```dotenv
ENV_AUTH_ISSUER=identity
ENV_AUTH_ORGANIZATION__ID=<organization served by this database>
ENV_AUTH_PUBLIC__KEYS={"key-1":"<raw URL-base64 Ed25519 public key>"}
```

Public keys must come from a trusted Identity configuration. Key IDs are preserved
exactly in the JSON map. Restart with the new key set for rotation, retaining old
keys until their tokens expire. Invalid/missing configuration prevents API startup;
migrations do not require authentication configuration. No private signing key is
needed. Helm deployments supply these variables through `envConfig` or
`secretKeyRef` before rolling out this version.

The database and Redis cache are scoped to one configured organization. Deploy
separate database/cache instances for different organizations; this service does
not implement shared database multitenancy. Internal callers, including appointment
and notification, must also send an appropriate Bearer token.

Run checks with `cd customer && go test ./... && go vet ./...`.
Build the container from the repository root:
`docker build -f customer/Dockerfile .` (the build includes the sibling core module).
