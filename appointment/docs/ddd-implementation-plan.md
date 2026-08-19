# Appointment DDD implementation plan

## Macro step 1: align notification recipient model

Status: completed

Goal: keep `AppointmentNotification` appointment-owned, but avoid hard-coding `customer_id`.

Tasks:

- Replace `appointment_notifications.customer_id` with `recipient_type` and `recipient_id`.
- Update sqlc schema and queries.
- Update legacy backfill command.
- Update repository writes.
- Update `domain/v2.AppointmentNotification`.
- Keep runtime behavior equivalent: current recipient is always `customer`.

## Macro step 2: make appointment notification tracking primary

Status: completed

Goal: remove runtime dependency on `pending_notifications`.

Tasks:

- Read pending outcomes from `appointment_notifications`.
- Mark sent/failed on `appointment_notifications`.
- Stop writing `pending_notifications` in runtime.
- Keep `pending_notifications` only as legacy input for manual backfill.
- Add tests for confirmation/rescheduled/reminder outcome behavior.

Progress:

- Runtime outcome lookup now reads from `appointment_notifications`.
- Runtime notification tracking writes only to `appointment_notifications`.
- Runtime application vocabulary now uses `AppointmentNotificationTracking`, not `PendingNotification`.

## Macro step 3: make appointment reminder state primary

Status: completed

Goal: stop using reminder columns on `agenda_events` for runtime behavior.

Tasks:

- Load reminder status from `appointment_reminders`.
- Write reminder transitions directly to `appointment_reminders`.
- Keep legacy columns only until cleanup migration.
- Add tests for schedule/send_requested/sent/failed/deleted transitions.

Progress:

- Repository reads reminder state from `appointment_reminders`.
- Reminder transitions now save through `SaveAppointmentReminder`, without rewriting the whole legacy `agenda_events` row.

## Macro step 4: introduce v2 application commands

Status: completed

Goal: move application code away from generic `CreateAgenda`.

Tasks:

- Add `CreateAppointment`.
- Add `CreateManualEvent`.
- Add `CreateTimeBlock`.
- Add `RescheduleAppointment`.
- Add `CancelAppointment`.
- Keep HTTP legacy as adapter.

Progress:

- Added `appointment/internal/application/v2`.
- Added `CreateAppointment`, `CreateManualEvent`, and `CreateTimeBlock` commands.
- Consolidated calendar use cases in one `CalendarService`.
- `CalendarEvent` is the single v2 aggregate root; `Appointment`, `ManualEvent`, and `TimeBlock` are mutually exclusive details.
- Create commands now build a `CalendarEvent` through typed factories (`NewAppointmentEvent`, `NewManualCalendarEvent`, `NewTimeBlockCalendarEvent`).
- `CalendarEvent` owns reschedule/cancel/update behavior and collects generic calendar lifecycle events (`CalendarEventCreated`, `CalendarEventRescheduled`, `CalendarEventCanceled`).
- Repository access is uniform through `FindCalendarEvent` and `SaveCalendarEvent`; subtype table writes and lifecycle outbox persistence stay inside Postgres infrastructure.
- HTTP legacy remains untouched.

## Macro step 5: introduce v2 repositories

Status: completed

Goal: persist and load domain aggregates directly, without using legacy `domain.AgendaEvent` as the main model.

Tasks:

- Add repository methods for `domain/v2.Appointment`.
- Add repository methods for `domain/v2.ManualEvent`.
- Add calendar read model queries.
- Move lifecycle/reminder code to v2 aggregate methods.

Progress:

- Added initial Postgres repository methods for saving v2 appointment, manual event, and time block calendar aggregates.
- Added SQLC query for saving v2 agenda event base fields while legacy columns still exist.
- `Appointment` no longer contains reminder state; reminder persistence is exposed separately.
- Removed the separate appointment identity from the v2 model: `appointments.agenda_event_id` is the primary key and child appointment tables reference that value.
- Creating an appointment now persists its pending reminder in the same transaction as the calendar event and lifecycle outbox message.
- New `CalendarEvent*` lifecycle messages are resolved through the v2 aggregate before dispatch: only appointment events reach the appointment reminder/notification handler.
- Legacy `AgendaEvent*` messages still bypass the v2 lookup while the old queue is being drained.
- Calendar GET/LIST now use a read projection that composes the `CalendarEvent` aggregate with its independent appointment reminder state.
- The reminder projection exposes pending, scheduling, request, delivery and failure timestamps without adding reminder state to `CalendarEvent`.
- Manual reminder resend now publishes the notification request, tracks its outcome and marks the reminder `send_requested` in one transaction, with a caller-provided idempotency key when available.
- `AppointmentLifecycleService` now handles new lifecycle messages, River due jobs, resend requests and notification outcomes entirely through v2 aggregates and projections.
- River scheduling/cancellation, reminder state, notification tracking and outbox publication share the transaction bound to `ContextDB`.
- River uses a dedicated producer client for transactional inserts and a separate runtime client for worker registration, avoiding a DI dependency cycle.
- The River worker and notification outcome consumer no longer depend on legacy `domain.AgendaEvent`.
