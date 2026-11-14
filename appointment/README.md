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
ENV_AUTH_PUBLIC__KEYS={"key-1":"<raw URL-base64 Ed25519 public key>"}
```

Use trusted Identity public keys and retain previous keys until their issued
tokens have expired during key rotation. Invalid or missing authentication
configuration prevents HTTP API startup; migrations do not require it.

Run checks with `cd appointment && go test ./... && go vet ./...`. Build the
container from the repository root: `docker build -f appointment/Dockerfile .`.
