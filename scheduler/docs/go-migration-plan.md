# Scheduler Go Migration Plan

## Summary

Migrate `scheduler` from Kotlin/Spring Boot to Go while preserving the existing functional contract:

- keep the current Scheduler HTTP API behavior;
- keep RabbitMQ as the delivery transport;
- replace Redis scheduler storage with Postgres;
- use Gin for HTTP, Zap for structured logging, OpenTelemetry for observability, Koanf for app configuration and environment variables, and Mage for cross-platform build/test/run tasks;
- use Cobra for the executable command surface:
  - `app` starts the HTTP server and scheduler runtime;
  - `migrate` applies Postgres schema migrations via `golang-migrate`;
  - `migrate-old` copies pending schedules from Redis into Postgres for continuity during cutover.

The migration should keep the service externally compatible for appointment-service and any other current clients.

## Current Behavior To Preserve

The current Scheduler API is defined in `scheduler/api-spec/openapi.yaml`:

- `PUT /schedules/{scheduleId}`
  - accepts `scheduleAt`, `route`, and `data`;
  - stores or replaces a scheduled job;
  - returns `202 Accepted` with the schedule id.
- `DELETE /schedules/{scheduleId}`
  - removes the scheduled job;
  - returns `204 No Content` on successful deletion.

The current domain model is:

- `ScheduleId`: string UUID;
- `ScheduleMeta`: `route` plus arbitrary JSON-like `data`;
- `ScheduleJob`: id, metadata, and `scheduleAt`.

The current runtime behavior is:

- poll due jobs periodically;
- lease polled jobs for a short TTL so multiple scheduler replicas do not deliver the same job concurrently;
- publish each due job to RabbitMQ;
- acknowledge a job only after successful publish;
- remove acknowledged jobs from storage.

The current RabbitMQ behavior must remain compatible:

- exchange: default exchange `""`;
- routing key: `job.meta.route`;
- body: JSON serialization of `job.meta.data`;
- content type: `application/json`.

## Target Structure

Replace the Gradle/Spring scheduler implementation with a Go module under `scheduler/`.

Recommended layout:

```text
scheduler/
├── api/
│   ├── openapi.yaml
│   └── oapi-codegen.yaml
├── cmd/
│   └── scheduler/
│       └── main.go
├── internal/
│   ├── app/
│   ├── application/
│   ├── config/
│   ├── container/
│   ├── domain/
│   ├── infra/
│   │   ├── postgres/
│   │   ├── rabbitmq/
│   │   └── redislegacy/
│   ├── port/
│   │   └── http/
│   └── runtime/
├── migrations/
├── magefile.go
├── Dockerfile
├── go.mod
└── go.sum
```

Use `oapi-codegen` to generate Gin server interfaces from the existing OpenAPI spec, following the same API-first direction used elsewhere in the repository.

Core stack:

- Cobra for CLI commands;
- Koanf for configuration loading from defaults, optional files, and environment variables;
- Gin for HTTP routing;
- Zap for logging;
- OpenTelemetry for traces and metrics;
- pgx for Postgres access;
- golang-migrate for schema migrations;
- amqp091-go for RabbitMQ;
- go-redis only in the temporary legacy migration adapter used by `migrate-old`.


Build tool decision:

- use Mage instead of Make because the scheduler must support Windows developer workflows;
- bob is not selected for this repository because its own comparison notes no Windows support, while Mage keeps the build tasks Go-native and cross-platform;
- keep runtime operations in Cobra, and use Mage only as the developer build/test/task layer.
## Command Surface

Use Cobra as the process entrypoint. The binary should expose exactly these operational commands first:

```text
scheduler app
scheduler migrate
scheduler migrate-old
```

### `app`

Starts the normal service:

- load config with Koanf;
- initialize Zap logger;
- initialize OpenTelemetry;
- build the dependency container;
- connect to Postgres;
- connect to RabbitMQ;
- run the Gin HTTP server;
- run the scheduler polling runtime;
- support graceful shutdown for HTTP, polling, RabbitMQ, Postgres, and telemetry flush.

This is the command used by the main Kubernetes container.

### `migrate`

Applies Postgres schema migrations using `golang-migrate`.

Expected behavior:

- load config with Koanf;
- read `POSTGRES_DSN` from typed config;
- load migrations from the embedded or filesystem `migrations/` directory;
- run `up`;
- treat `migrate.ErrNoChange` as success;
- fail fast on any other migration error.

This command should be used by the scheduler Helm chart as an init container before `app` starts.

### `migrate-old`

Copies pending Redis schedules into Postgres so the migration does not lose continuity.

Expected behavior:

- load config with Koanf;
- connect to Redis using legacy scheduler Redis config;
- connect to Postgres using typed Postgres config;
- read the legacy Redis sorted set `${SCHEDULER_NAME}-clock`;
- for each schedule id in the sorted set, read the JSON payload from the Redis string key;
- deserialize the legacy Kotlin/Spring `ScheduleJob` payload;
- insert into Postgres using idempotent upsert semantics;
- preserve `id`, `route`, `data`, and `scheduleAt`;
- do not delete Redis data by default;
- log copied, skipped, failed, and already-existing counts;
- support a dry-run option if practical.

This command can be run as a one-shot job or init container during cutover. It should be safe to run more than once.

## Postgres Storage Design

Create a durable schedule table that replaces the Redis sorted-set plus value-key model.

Initial schema:

```sql
CREATE TABLE schedule_jobs (
  id UUID PRIMARY KEY,
  scheduler_name TEXT NOT NULL,
  route TEXT NOT NULL,
  payload JSONB NOT NULL,
  schedule_at TIMESTAMPTZ NOT NULL,
  leased_until TIMESTAMPTZ,
  lease_owner TEXT,
  attempts INT NOT NULL DEFAULT 0,
  last_error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX schedule_jobs_due_idx
ON schedule_jobs (scheduler_name, schedule_at)
WHERE leased_until IS NULL;

CREATE INDEX schedule_jobs_lease_idx
ON schedule_jobs (scheduler_name, leased_until);
```

Repository behavior:

- `Save(job)` performs an upsert by `id`, replacing route, payload, schedule time, and clearing any stale lease.
- `Delete(id)` removes the job. Prefer idempotent delete and return success even when the row does not exist, unless the API contract is tightened to enforce `404`.
- `PollDue(now, batchSize, leaseTTL, leaseOwner)` atomically selects due jobs and leases them.
- `Ack(id)` deletes the row.
- `Nack(id, reason)` records the error and releases or shortens the lease so the job can be retried.

Use `FOR UPDATE SKIP LOCKED` for safe concurrent polling:

```sql
WITH due AS (
  SELECT id
  FROM schedule_jobs
  WHERE scheduler_name = $1
    AND schedule_at <= $2
    AND (leased_until IS NULL OR leased_until <= $2)
  ORDER BY schedule_at ASC
  LIMIT $3
  FOR UPDATE SKIP LOCKED
)
UPDATE schedule_jobs j
SET leased_until = $4,
    lease_owner = $5,
    attempts = attempts + 1,
    updated_at = now()
FROM due
WHERE j.id = due.id
RETURNING j.id, j.scheduler_name, j.route, j.payload, j.schedule_at, j.leased_until, j.lease_owner, j.attempts, j.last_error;
```

This preserves the important Redis lease behavior: if a process dies after polling but before acking, the job becomes eligible again after `leased_until`.

## Dependency Container

Create a simple dependency container instead of wiring dependencies directly in Cobra commands.

Suggested package:

```text
internal/container
```

Suggested responsibilities:

- receive the loaded `config.Config`;
- create and own the Zap logger;
- initialize OpenTelemetry providers and expose a shutdown function;
- create Postgres connection pool;
- create RabbitMQ connection/channel or publisher;
- create repositories and adapters;
- create application services;
- create the scheduler runtime;
- create the Gin router/server;
- expose cleanup through `Close(ctx)` or equivalent.

The container should stay explicit and simple. Avoid a reflection-based DI framework. A small struct is enough:

```go
type Container struct {
    Config    config.Config
    Logger    *zap.Logger
    DB        *pgxpool.Pool
    Publisher *rabbitmq.Publisher
    Store     *postgres.JobRepository
    Scheduler *application.SchedulerService
    Runtime   *runtime.Runtime
    HTTP      *http.Server
}
```

Command usage:

- `app` builds the full runtime container and starts HTTP plus polling;
- `migrate` can use a smaller migration container or direct config plus logger plus Postgres DSN;
- `migrate-old` builds only config, logger, Redis legacy client, and Postgres repository.

Keep lifecycle ownership clear: anything opened by the container must be closed by the container.

## Runtime Design

Implement a polling runtime with `time.Ticker`.

For each tick:

1. call `PollDue`;
2. for each returned job, publish to RabbitMQ;
3. call `Ack` only after successful publish;
4. call `Nack` after publish failure;
5. continue processing the batch even if one job fails.

The runtime should accept:

- scheduler name;
- polling interval;
- peek batch size;
- lease TTL;
- lease owner, generated per process instance.

The runtime must stop cleanly when the `app` command receives `SIGINT` or `SIGTERM`.

## RabbitMQ Adapter

Use `github.com/rabbitmq/amqp091-go`.

Publish messages with:

- exchange configured by `RABBIT_EXCHANGE`, defaulting to `""`;
- routing key equal to the stored `route`;
- JSON body equal to stored `payload`;
- content type `application/json`;
- delivery mode persistent if queue durability expects it.

The adapter should surface publish errors to the runtime so jobs are not acked prematurely.

## HTTP API

Use Gin and generated OpenAPI handlers.

Handler behavior:

- parse UUID path parameter;
- validate request body;
- preserve arbitrary JSON `data`;
- convert OpenAPI DTO to domain command;
- call scheduler application service;
- return compatible status codes and response bodies.

Health endpoints:

- expose `/health`;
- expose `/actuator/health/liveness` and `/actuator/health/readiness` as compatibility endpoints so the current Helm probes can stay stable during the first migration.

## Configuration

Use Koanf for all app configuration and environment variable handling.

Configuration loading order:

1. defaults from code;
2. optional config file, if a path is provided;
3. environment variables;
4. Cobra flags only for command-specific overrides such as dry-run.

Use a typed `config.Config` struct so the rest of the service does not read Koanf directly.

Suggested shape:

```go
type Config struct {
    Server    ServerConfig
    Scheduler SchedulerConfig
    Postgres  PostgresConfig
    RabbitMQ  RabbitMQConfig
    Redis     RedisLegacyConfig
    Otel      OtelConfig
    Log       LogConfig
}
```

Suggested environment names:

```text
SERVER_PORT=8080

SCHEDULER_NAME=reminders
SCHEDULER_POLLING_INTERVAL=60s
SCHEDULER_PEEK_LEASE_TTL=30s
SCHEDULER_PEEK_BATCH_SIZE=50

POSTGRES_DSN=postgres://...

RABBIT_HOST=localhost
RABBIT_PORT=5672
RABBIT_USERNAME=guest
RABBIT_PASSWORD=guest
RABBIT_EXCHANGE=

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

OTEL_COLLECTOR_GRPC_ENDPOINT=http://localhost:4317
LOG_LEVEL=info
```

During migration, support legacy Redis env vars only for `migrate-old`.

Normalize RabbitMQ username configuration. Current scheduler config uses `RABBIT_USERNAME`, while `scheduler/helm/values.yaml` currently uses `RABBIT_USER`. The Go service should either support both during transition or update Helm to use `RABBIT_USERNAME`.

## Observability

Use Zap for structured logs.

Log at minimum:

- service startup and shutdown;
- migration start/end;
- Redis legacy migration counts;
- schedule creation and deletion;
- poll batch size and duration;
- RabbitMQ publish success/failure;
- job ack/nack.

Include fields:

- `schedule_id`;
- `scheduler_name`;
- `route`;
- `attempts`;
- `lease_owner`;
- `error`.

Use OpenTelemetry for:

- incoming HTTP requests;
- Postgres operations;
- RabbitMQ publish spans;
- scheduler polling spans.

Expose metrics or instruments for:

- jobs scheduled;
- jobs deleted;
- jobs polled;
- jobs published;
- publish failures;
- nack count;
- polling duration;
- due job lag.

## Build Tasks

Add a scheduler-local Mage file with cross-platform targets:

```text
deps
generate
build
run
test
test-coverage
lint
docker-build
migrate
migrate-old
clean
```

Expected command mapping:

- `mage run` runs `go run ./cmd/scheduler app`;
- `mage migrate` runs `go run ./cmd/scheduler migrate`;
- `mage migrateOld` runs `go run ./cmd/scheduler migrate-old`;
- `mage generate` runs OpenAPI generation;
- `mage test` runs `go test ./...`.

## Docker And Helm

Replace the scheduler Dockerfile with a Go multi-stage build.

Main container command:

```text
scheduler app
```

Add an init container for schema migration:

```text
scheduler migrate
```

Add a one-time migration path for Redis continuity:

```text
scheduler migrate-old
```

The `migrate-old` command should not be permanently enabled on every deploy unless it is proven harmless and idempotent in the target environment. Prefer a dedicated Kubernetes Job or a temporary init container during the cutover window.

Helm changes:

- remove Redis env vars from the normal `app` container after cutover;
- keep Redis env vars only for the temporary `migrate-old` job/init container;
- add `POSTGRES_DSN` or equivalent secret/config references;
- keep RabbitMQ env vars;
- keep existing probe paths if compatibility health endpoints are implemented.

Docker Compose changes:

- add a Postgres service for local development;
- keep Redis during migration testing;
- keep RabbitMQ;
- add scheduler service only after the Go service can run locally.

## Cutover Plan

Preferred low-risk cutover:

1. Deploy Postgres and run `scheduler migrate`.
2. Deploy the Go scheduler image without routing traffic yet, or keep replicas at zero.
3. Stop or scale down the old Kotlin/Spring scheduler so it no longer polls and publishes due jobs.
4. Run `scheduler migrate-old` to copy all pending Redis schedules into Postgres.
5. Start the Go scheduler with `scheduler app`.
6. Verify new schedules are written to Postgres.
7. Verify due schedules are published to RabbitMQ and acknowledged from Postgres.
8. Keep Redis data temporarily for rollback.
9. Remove Redis dependency from scheduler deployment after the cutover is stable.

Rollback strategy:

- if the Go scheduler fails before Redis cleanup, scale down Go scheduler and scale up the old Kotlin/Spring scheduler;
- because `migrate-old` does not delete Redis data by default, pending legacy schedules remain available to the old service;
- any schedules created only in Postgres after cutover need a rollback decision before reverting.

## Test Plan

Unit tests:

- schedule create maps id, route, payload, and schedule time correctly;
- delete is idempotent or returns not found according to final API decision;
- Koanf config maps defaults, environment variables, and durations correctly;
- the dependency container wires expected services and closes owned resources;
- runtime acks only after successful publish;
- runtime nacks after publish failure;
- `migrate` treats no-op migrations as success;
- `migrate-old` maps legacy Redis JSON into the Postgres model.

Postgres repository tests:

- save inserts a new job;
- save updates an existing job with same id;
- poll returns only due jobs;
- poll respects batch size;
- concurrent pollers do not receive the same job;
- leased jobs are hidden until lease expiry;
- expired leases become pollable again;
- ack deletes the job;
- nack records error and makes retry possible.

HTTP tests:

- `PUT /schedules/{scheduleId}` returns `202`;
- invalid UUID returns `400`;
- invalid payload returns `400`;
- `DELETE /schedules/{scheduleId}` returns the chosen compatible status;
- health endpoints return success when dependencies are ready.

RabbitMQ tests:

- publish uses default exchange;
- routing key is the schedule route;
- body is exactly the schedule payload JSON;
- publish failure propagates to runtime.

Integration tests:

- with Postgres and RabbitMQ, a schedule in the past is published within one polling interval;
- killing the process after lease but before ack causes delivery retry after lease TTL;
- running `migrate-old` twice does not duplicate jobs;
- migrated Redis jobs are delivered by the Go scheduler.

## Assumptions

- The Scheduler public API remains unchanged.
- RabbitMQ remains the delivery mechanism.
- Postgres becomes authoritative for new scheduler state.
- Redis is used only as legacy source data during `migrate-old`.
- Koanf is the only configuration access path used by application code.
- DI remains manual through a simple container; no runtime reflection DI framework is introduced.
- `migrate-old` must be idempotent and non-destructive by default.
- `nack` can improve on the current Kotlin TODO as long as at-least-once delivery is preserved.
- A brief cutover window is acceptable. Full zero-downtime migration would require additional dual-write or traffic routing design.