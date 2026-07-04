# Appointment Service Go Migration Plan

## Summary

Introduce a Go implementation beside the Kotlin service. The first target is functional parity for agenda activities and services, with PostgreSQL as source of truth and MongoDB backfill. Queue-driven reminder/notification flows are represented through outbox-ready tables and adapter boundaries.

## Implementation Changes

- Cobra commands: `app`, `migrate`, `backfill`.
- Koanf env config using `__` separators.
- Gin HTTP server implementing generated OpenAPI `ServerInterface`.
- PostgreSQL schema for `agenda_events`, `appointment_services`, `pending_notifications`, and `outbox_messages`.
- MongoDB backfill for legacy `agenda` and `services` collections.
- DDD package split: `domain`, `application`, `infra`, `port/http`.
- Mage targets for `generate`, `build`, `test`, `lint`, `check`.

## Test Plan

- `go test ./...`
- `mage check`
- migration smoke test against Postgres
- backfill against copied Mongo data
- HTTP parity checks for activity and service CRUD/search
- end-to-end reminder and notification confirmation checks before cutover

## Assumptions

- Existing upstream clients use the OpenAPI contract rather than Kotlin-specific serialization.
- Initial cutover can use Postgres outbox plus the existing outbox-forwarder service.
- Redis notification correlation can be represented as `pending_notifications` in Postgres.
