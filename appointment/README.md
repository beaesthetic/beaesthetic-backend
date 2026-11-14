# Appointment service

The public `/v1` API verifies Identity `v4.public` PASETO tokens through
`core/auth`. Calendar events and service catalogue reads require
`appointment.read`; creations, updates, cancellations, and reminder resends
require `appointment.write`. `/health` remains public.

Missing or invalid tokens return `401` with a Bearer challenge. A membership
token without the required permission, or a token with a non-membership scope,
returns `403` before request handling.

Set these values through the environment, an `--env-file`, or Helm before
starting the HTTP API:

```dotenv
ENV_AUTH_ISSUER=identity
ENV_AUTH_PUBLIC__KEYS__URL=http://identity-service:8080/v1/internal/keys
ENV_AUTH_INTERNAL__API__KEY=<shared Identity internal API key>
```

The service retrieves the trusted Identity key set at startup. Alternatively,
set `ENV_AUTH_PUBLIC__KEYS` to a JSON `kid -> public_key_b64` map. Invalid or
missing authentication configuration prevents HTTP API startup; migrations do
not require it.

Run checks with `cd appointment && go test ./... && go vet ./...`. Build the
container from the repository root: `docker build -f appointment/Dockerfile .`.
