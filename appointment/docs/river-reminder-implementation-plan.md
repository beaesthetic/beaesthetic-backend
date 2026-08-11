# River Reminder Implementation Plan

Implementation status: completed for the v2 runtime path. The legacy scheduler queue remains enabled only for migration compatibility.

## Summary

Move appointment reminder scheduling from the external `scheduler` service to River jobs owned by the `appointment` service.

The goal is not to replace RabbitMQ/outbox for cross-service communication. River should only replace delayed execution for appointment reminders.

Current delayed reminder flow:

1. Appointment lifecycle handler computes `sendAt`.
2. Appointment calls scheduler REST API.
3. Scheduler stores `schedule_jobs`.
4. Scheduler publishes a due event to RabbitMQ.
5. Appointment consumes the due event from `SchedulerQueueConsumer`.
6. Appointment publishes a customer notification request through outbox.

Target flow:

1. Appointment lifecycle handler computes `sendAt`.
2. Appointment inserts a River job scheduled for `sendAt`.
3. River worker runs inside appointment when due.
4. Worker reloads the v2 calendar event projection and validates it is still sendable.
5. Worker publishes a customer notification request through the existing outbox sender.

## Source References

River is a Go job queue backed by Postgres. It supports scheduled jobs, transactional job insertion, typed workers, retries, unique jobs, and pgx integration.

Relevant upstream docs:

- https://riverqueue.com/
- https://github.com/riverqueue/river

## Scope

In scope:

- Add River to appointment.
- Add River migrations to appointment DB migrations.
- Add a typed `SendAppointmentReminder` job.
- Replace the HTTP scheduler adapter with a River adapter behind the existing `application.ReminderScheduler` interface.
- Add a River worker for due reminders.
- Keep existing notification outbox flow unchanged.

Out of scope for the first implementation:

- Removing the `scheduler` service from the repository.
- Removing scheduler Helm/release pipeline.
- Migrating old scheduled jobs from scheduler DB.
- Changing reminder business rules.
- Moving all asynchronous work from RabbitMQ to River.

## Current Appointment Integration Points

The useful boundary already exists:

```go
type ReminderScheduler interface {
    ScheduleReminder(ctx context.Context, agendaEvent *domain.AgendaEvent, sendAt time.Time) error
    UnscheduleReminder(ctx context.Context, eventID string) error
}
```

Current implementation:

- DI method: `GetReminderScheduler`
- River adapter: `appointment/internal/infra/jobs`
- scheduler creation runs inside the same Postgres transaction as reminder status, notification tracking and outbox writes.
- `AppointmentLifecycleService` is the application boundary used by new lifecycle messages, River workers and notification outcomes.

Current due-event consumer:

- `appointment/internal/infra/messaging/SchedulerQueueConsumer`
- reads RabbitMQ message `{ "eventId": "..." }`
- sends reminder notification
- tracks pending notification
- calls `ProcessReminderTimesUp`

During the migration window:

- `ReminderScheduler` becomes a River job scheduler.
- `SchedulerQueueConsumer` remains only to drain old RabbitMQ due events during migration.
- River uses the v2 lifecycle service.
- The legacy consumer keeps its old sender only to drain already-published messages.

## Proposed River Job

Job kind:

```text
appointment.send_reminder
```

Args:

```go
type SendAppointmentReminderArgs struct {
    EventID          string    `json:"eventId"`
    ExpectedStartAt  time.Time `json:"expectedStartAt"`
    ExpectedVersion  int       `json:"expectedVersion"`
}
```

Minimum first version:

```go
type SendAppointmentReminderArgs struct {
    EventID string `json:"eventId"`
}
```

Recommended first version includes `ExpectedStartAt` or `ExpectedVersion` to prevent stale jobs from sending reminders after reschedules. The repository already stores `agenda_events.version`, so `ExpectedVersion` is available if exposed through the domain model.

## Stale Job Protection

River scheduled jobs can be delayed; the appointment may be changed before the job runs. The worker must reload the appointment and verify:

- event still exists
- event is not cancelled
- reminder status is still scheduled or sendable
- event start time/version still matches what was scheduled
- event has not already requested or confirmed reminder sending

Recommended behavior:

- stale job: return nil, do not retry
- missing event: return nil, do not retry
- cancelled event: return nil, do not retry
- notification sender error: return error, let River retry
- database/outbox error: return error, let River retry

## Uniqueness And Reschedule

Use one pending reminder job per appointment.

Desired uniqueness key:

```text
appointment-reminder:{eventID}
```

Scheduling a reminder should replace or supersede the previous pending reminder for the same event.

There are two implementation options:

1. Use River unique jobs.
   - Prefer this if River supports the needed uniqueness mode for scheduled jobs in the open-source version.
   - Unique key is based on `EventID`.

2. Allow multiple River jobs but validate in worker.
   - Simpler if unique replacement is awkward.
   - Each job carries `ExpectedStartAt` or `ExpectedVersion`.
   - Old jobs no-op when they run.

Recommended first implementation:

- use unique jobs if straightforward
- still validate event state in the worker

The validation is required either way.

## Transaction Boundary

The best target is to insert River jobs in the same Postgres transaction as appointment changes.

Current lifecycle flow does not do this perfectly:

- `SaveAgendaEvent` publishes lifecycle events to outbox.
- `AppointmentLifecycleConsumer` later handles those events and schedules reminders.
- Scheduling currently happens after the original agenda transaction.

First River implementation can keep that behavior:

- lifecycle event is consumed
- lifecycle handler schedules River reminder job
- lifecycle handler marks reminder scheduled

This matches current reliability characteristics and is a smaller migration.

Second pass improvement:

- move reminder job insertion closer to the agenda transaction
- or make lifecycle handling transactional across status update and River insertion

Do not attempt the second pass in the first migration unless the River transaction API fits cleanly with `ContextDB`.

## Implementation Steps

### Phase 1: Add River Infrastructure

1. Add River dependencies to `appointment/go.mod`.
   - River client.
   - River pgx driver.

2. Add River migrations to appointment migrations.
   - Use River's official migration mechanism or include generated SQL migrations.
   - Keep migrations ordered after the existing appointment migrations.

3. Add config.
   - enable/disable River worker
   - River queue name
   - worker concurrency

Suggested env:

```text
RIVER__ENABLED=true
RIVER__QUEUE=appointment_reminders
RIVER__WORKERS=5
```

### Phase 2: Add Job Scheduler Adapter

Create a new adapter, for example:

```text
appointment/internal/infra/jobs/reminder_scheduler.go
```

It implements:

```go
application.ReminderScheduler
```

Methods:

- `ScheduleReminder(ctx, eventID, sendAt)`
- `UnscheduleReminder(ctx, eventID)`

`ScheduleReminder` inserts a River job scheduled at `sendAt`.

`UnscheduleReminder` should cancel/delete the pending River job for the event if possible. If cancellation is not straightforward, record enough state so a future stale job no-ops.

### Phase 3: Add Reminder Worker

Create a River worker, for example:

```text
appointment/internal/infra/jobs/send_appointment_reminder_worker.go
```

The worker should reuse the existing dependencies:

- `AppointmentService`
- `NotificationSender`
- logger

Worker flow:

1. Load agenda event by `EventID`.
2. Return nil if not found or cancelled.
3. Return nil if reminder is no longer pending/scheduled.
4. Validate expected start/version if present.
5. Send appointment reminder with `NotificationSender.SendAppointmentReminder`.
6. Track pending notification.
7. Call `ProcessReminderTimesUp`.

This logic should be extracted from `SchedulerQueueConsumer.Process` into a reusable application service/helper so RabbitMQ and River paths cannot drift during migration.

### Phase 4: Wire River Into App Runtime

In DI:

- add River client singleton
- add River worker registration
- add River start/stop lifecycle
- `GetReminderScheduler` always returns the River adapter

In `appointment/cmd/appointment/root.go`:

- start River client in the errgroup
- stop it on shutdown
- keep existing RabbitMQ consumers during migration

### Phase 5: Migration Compatibility

During transition:

- keep `SchedulerQueueConsumer` running
- remove the HTTP scheduler client from appointment
- keep `ENV_RABBITMQ_SCHEDULER__QUEUE` only while draining old scheduler due events

### Phase 6: Cutover

1. Run appointment migrations, including River migrations.
2. Deploy appointment with River scheduler enabled by default.
3. Monitor reminder scheduling and River jobs.
4. Keep scheduler due-event consumer for one deploy window.
5. Remove scheduler queue consumer after confidence.
6. Decommission scheduler service if no other service uses it.

## Backfill Strategy

No automatic data migration in the first implementation.

For existing future appointments:

- run existing `schedule-future-reminders`
- it schedules River jobs

For old jobs already stored in scheduler DB:

- allow them to fire during migration window
- or explicitly clear them after River jobs are scheduled

Avoid dual sending by worker validation:

- if reminder status is already `SEND_REQUESTED` or `SENT`, the River worker no-ops
- if old scheduler due event arrives after River sent, the old consumer should also no-op after we add the same validation

## Required Code Changes

Expected files to add:

- `appointment/internal/infra/jobs/reminder_scheduler.go`
- `appointment/internal/infra/jobs/send_appointment_reminder_worker.go`
- `appointment/internal/infra/jobs/client.go`
- River migration files under `appointment/migrations`

Expected files to modify:

- `appointment/internal/config/config.go`
- `appointment/cmd/di/infra.go`
- `appointment/cmd/di/services.go`
- `appointment/cmd/appointment/root.go`
- `appointment/internal/infra/messaging/scheduler_queue_consumer.go`
- `appointment/internal/application/lifecycle.go` only if the interface needs richer args

Expected files to remove later:

- `appointment/internal/infra/messaging/scheduler_queue_consumer.go`

## Test Plan

Unit tests:

- scheduler adapter inserts scheduled job at computed `sendAt`
- stale worker job no-ops
- cancelled appointment no-ops
- missing appointment no-ops
- valid worker job sends notification and tracks pending notification
- notification sender error returns error for retry

Integration tests:

- River job scheduled in Postgres
- River worker processes due reminder
- rescheduled appointment does not send old reminder
- manual `schedule-future-reminders` schedules River jobs

Regression tests:

- existing lifecycle tests still pass
- existing scheduler queue consumer tests pass during migration
- notification confirmation flow unchanged

## Risks

- Duplicate reminders during cutover if old scheduler jobs and River jobs coexist.
- River migration/version management needs to fit existing golang-migrate setup.
- Unscheduling unique scheduled jobs may require careful River API usage.
- If future services need generic scheduling, removing scheduler may be premature.

## Recommendation

Implement River in appointment behind the existing `ReminderScheduler` interface and remove the HTTP scheduler adapter from appointment. Keep only the legacy due-event consumer temporarily to drain old scheduler jobs.

After River is proven in appointment:

- remove scheduler due-event consumer
- decommission scheduler service only if no other domain needs generic delayed jobs
