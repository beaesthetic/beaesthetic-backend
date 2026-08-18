# Appointment Service - Current State Memo

## Context

The current appointment service is implemented with Quarkus + Kotlin. It owns agenda activities, appointment services, reminder orchestration, notification composition, and confirmation correlation.

## Runtime Entrypoints

- Calendar API: the `calendars/v1` protobuf contract, served by the calendar-event HTTP handlers.
- Scheduler queue consumer: consumes reminder times-up events from the scheduler queue.
- Notification confirmation consumer: consumes notification confirmation events and correlates `notificationId` back to agenda reminders.
- Failed reminder monitor: periodically recovers reminders stuck in failed/in-progress states.

## Current Storage

- MongoDB collection `agenda` stores agenda events with optimistic versioning, reminder state, attendee data, and polymorphic event data.
- MongoDB collection `services` stores bookable services with search grams.
- Redis stores pending notification correlation: `notificationId -> agendaEventId + notificationType`.

## Core Functional Flows

1. Create/update/delete agenda event through REST.
2. Persist agenda event in MongoDB.
3. Publish lifecycle events internally.
4. Policies react to lifecycle events:
   - schedule/delete reminders through scheduler REST API
   - send confirmation/reschedule notifications through notification REST API
5. Scheduler queue tells appointment when reminder time is due.
6. Appointment sends reminder notification and tracks `notificationId` correlation in Redis.
7. Notification confirmation queue marks reminder as sent and removes Redis correlation.

## Migration Constraints

- The calendar API contract is defined by the `calendars/v1` protobuf messages.
- Mongo optimistic versioning maps to Postgres integer version checks.
- Save plus lifecycle publication should move to transactional outbox.
- Redis correlation should move to Postgres or a durable table to avoid split storage.
- Reminder and notification payload compatibility must be preserved across scheduler/notification services.
